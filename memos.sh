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
                        "id": .id
                    },
                },
                {
                "type": "run",
                "title": "edit memo",
                "command": "edit-memo",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                    "id": .id
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
    id=$(echo "$1" | jq -r '.params.id')
    content=$(echo "$1" | jq -r '.params.content')
    echo "$1" | jq -r '.params.content' > /tmp/$id.md
    file="/tmp/$id.md"
    codeblock=$(echo "$1" | jq -r '.params.codeblock')
    
    jq -n --arg content "$content" --arg codeblock "$codeblock"  --arg file "$file" '{
        "markdown": $content,
        "actions": [{
            title: "Copy to clipboard",
            text: $codeblock,
            type: "copy",
            exit: false,
            },
            {
            type: "edit",
            title: "Edit",
            path: $file,
            exit: false,
        }],
    }' 2> >(tee /dev/stderr) || { 
       echo "Error: jq failed to process JSON" >&2
       exit 2
    }
fi

if [ "$COMMAND" = "edit-memo" ]; then
    content=$(echo "$1" | jq -r '.params.content')
    codeblock=$(echo "$1" | jq -r '.params.codeblock')
    id=$(echo "$1" | jq -r '.params.id')
    echo $content > /tmp/${id}.md
    sunbeam edit /tmp/${id}.md
fi
