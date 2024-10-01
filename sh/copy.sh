#!/bin/bash

append() {
    local file_path="$1"
    echo -e "${file_path}:\n\`\`\`"
    cat "${file_path}" | sed 's/^\s+//' | sed 's/\s+$//'
    echo -e "\`\`\`"
}

export -f append

echo "Copying contents to clipboard..."
(
    tree

    find -type f -name 'go.mod' -exec bash -c '
        append "$1"
    ' _ {} \;

    find ./ -type f -name '*.go' -exec bash -c '
        append "$1"
    ' _ {} \;
) | xsel --clipboard

echo "Contents copied to clipboard."
