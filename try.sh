#!/bin/bash

# Define source and destination paths
SOURCE="/Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/starfishAI/aacurtor.drawio"
DESTINATION="/Users/john/Documents/projects/aa/python/starfish/starfish/examples/starfish.drawio"

# Copy the file
cp "$SOURCE" "$DESTINATION"

# Check if the copy was successful
if [ $? -eq 0 ]; then
    echo "File copied successfully."
else
    echo "Failed to copy the file."
fi