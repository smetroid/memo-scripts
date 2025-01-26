#!/bin/sh

#Surpresses errors from the shell, disable when debugging
#set -eu

OUTPUT="/tmp/memos.json"

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Memos",
        description: "Search memo code blocks",
        preferences: [
            {
                name: "memo_token",
                title: "Memo Personal Access Token",
                type: "string"
            },
            {
                name: "memo_url",
                title: "Memo API URL",
                type: "string"
            }
        ],
        commands: [
            {
                name: "memo-cmds",
                title: "Memo cmds blocks",
                mode: "filter"
            },
            {
                name: "memo-snippets",
                title: "Memo snippets blocks",
                mode: "filter"
            },
            {
                name: "memo-all",
                title: "ALL memos",
                mode: "filter"
            },
            {
                name: "run-command",
                title: "execute command",
                mode: "tty",
                exit: "true"
            },
            {
                name: "view-command",
                title: "view command",
                mode: "detail",
                exit: "false"
            },
            {
                name: "edit-memo",
                title: "edit memo",
                mode: "tty",
                exit: "false"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
FILTER=$(echo "$1" | jq -r '.command | split("-")[1]')
if [ "$COMMAND" = "memo-cmds" ]; then
    echo $(date) >>$OUTPUT
    MEMOS=$(~/projects/memo-scripts/get-memos -tags "${FILTER}")
    echo "Debug: MEMOS output:" >>$OUTPUT
    echo "$FILTER" >>$OUTPUT
    echo "$MEMOS" >>$OUTPUT
    # it seems to fail because get-memos is not fast enough
    #~/projects/memo-scripts/get-memos | tee ./debug_output.json | jq '{
    echo "$MEMOS" | jq '{
        "items": map({
            "title": .cmd,
            #"subtitle": .tags,
            "accessories": [.tags],
            "actions": [{
                "type": "run",
                "title": "Run Command",
                "command": "run-command",
                "params": {
                    "codeblock": .cmd,
                }
                },{
                "type": "run",
                "title": "View Command",
                "command": "view-command",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                },
            }]
        }),
        "actions": [{
          "title": "Refresh items",
          "type": "reload",
          "exit": "true"
      }]
  }'
    exit 0
fi

if [ "$COMMAND" = "memo-snippets" ]; then
    MEMOS=$(~/projects/memo-scripts/get-memos -tags "${FILTER}")
    # it seems to fail because get-memos is not fast enough
    #~/projects/memo-scripts/get-memos | tee ./debug_output.json | jq '{
    echo "$MEMOS" | jq '{
        "items": map({
            "title": .content,
            #"subtitle": .tags,
            "accessories": [.tags],
            "actions": [{
                "type": "run",
                "title": "view cmd ",
                "command": "view-command",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                },
            }]
        }),
        "actions": [{
          "title": "Refresh items",
          "type": "reload",
          "exit": "true"
      }]
  }'
    exit 0
fi

if [ "$COMMAND" = "memo-all" ]; then
    MEMOS=$(~/projects/memo-scripts/get-memos)
    # it seems to fail because get-memos is not fast enough
    #~/projects/memo-scripts/get-memos | tee ./debug_output.json | jq '{
    echo "$MEMOS" | jq '{
        "items": map({
            "title": .content,
            "subtitle": .cmd,
            "accessories": [.tags],
            "actions": [{
                "type": "run",
                "title": "view memo",
                "command": "view-command",
                    "params": {
                        "content": .content,
                        "codeblock": .cmd,
                        "name": .name
                    },
                },
                {
                "type": "run",
                "title": "edit memo",
                "command": "edit-memo",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                    "name": .name
                }
            }]
        }),
        "actions": [{
          "title": "Refresh items",
          "type": "reload",
          "exit": "true"
        }]
  }' 2> >(tee /dev/stderr) || { 
       echo "Error: jq failed to process JSON" >&2
       exit 2
    }
  exit 0
fi

if [ "$COMMAND" = "run-command" ]; then
    CMD=$(echo "$1" | jq -r '.params.codeblock')
    konsole -e bash -c "$CMD; exec bash"
elif [ "$COMMAND" = "view-command" ]; then
    content=$(echo "$1" | jq -r '.params.content')
    codeblock=$(echo "$1" | jq -r '.params.codeblock')
    name=$(echo "$1" | jq -r '.params.name')
    
    jq -n \
    --arg content "$content" \
    --arg codeblock "$codeblock"  \
    --arg name "$name" '{
        "markdown": $content,
        "actions": [{
            type: "copy",
            title: "Copy to clipboard",
            text: $codeblock,
            exit: false
        },
        {
            type: "run",
            title: "Edit",
            command: "edit-memo",
            params: {
                "content": $content,
                "codeblock": $codeblock,
                "name": $name
            }
        }],
    }' 2> >(tee /dev/stderr) || { 
        echo "Error: jq failed to process JSON" >&2
        exit 2
    }
fi

if [ "$COMMAND" = "edit-memo" ]; then
    # the name is in the form of "memos/id"
    if [ ! -d "/tmp/memos" ]; then
        mkdir /tmp/memos
    fi
    name=$(echo "$1" | jq -r '.params.name')
    TMP_FILE="/tmp/$name.md"
    content=$(echo "$1" | jq -r '.params.content')
    codeblock=$(echo "$1" | jq -r '.params.codeblock')
    name=$(echo "$1" | jq -r '.params.name')
    echo "$1" | jq -r '.params.content' > $TMP_FILE
    # Get the modification time before editing
    before_edit=$(stat -c %Y "$TMP_FILE")
    # Open the file in vim for editing
    vim $TMP_FILE
    # Get the modification time after editing
    after_edit=$(stat -c %Y "$TMP_FILE")

    if [ "$before_edit" -ne "$after_edit" ]; then
        STATUS=$(~/projects/memo-scripts/post-memo -update -name ${name})
        if [ $? -eq 0 ]; then
            echo "Success: updated memo"
            echo $STATUS
            read -p "Press enter to exit, and go back to the app"
            exit 0
        else
            echo "Error: failed to update memo"
            echo $STATUS
            read -p "Press enter to exit, and go back to the app"
            exit 1
        fi
    else
        echo "No changes made to the file."
        read -p "Press enter to exit, and go back to the app"
        exit 0
    fi
fi
