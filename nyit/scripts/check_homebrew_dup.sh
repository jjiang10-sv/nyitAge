#!/bin/bash

# Loop through all files in /usr/local/bin
for file in /usr/local/bin/*; do
    fname=$(basename "$file")
    homebrew_bin="/opt/homebrew/bin/$fname"
    if [ -e "$homebrew_bin" ]; then
        if [ -L "$file" ]; then
            # It's a symlink, check where it points
            target=$(readlink "$file")
            if [[ "$target" == "$homebrew_bin" ]]; then
                echo "$file is a symlink to $homebrew_bin"
            else
                echo "$file is a symlink, but not to $homebrew_bin (points to $target)"
            fi
        else
            echo "$file exists in both locations, but is not a symlink"
        fi
    fi
done