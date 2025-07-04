You can start Firefox from the command line on macOS using these commands:

## Basic Firefox Launch
```bash
# Start Firefox normally
/Applications/Firefox.app/Contents/MacOS/firefox

# Or if Firefox is in your PATH (after installation)
open -a Firefox
```

## Firefox with Specific Options
```bash
# Start Firefox with a new profile
/Applications/Firefox.app/Contents/MacOS/firefox -P

# Start Firefox with profile manager
/Applications/Firefox.app/Contents/MacOS/firefox -ProfileManager

# Start Firefox in private browsing mode
/Applications/Firefox.app/Contents/MacOS/firefox -private

# Start Firefox with a specific profile
/Applications/Firefox.app/Contents/MacOS/firefox -P "ProfileName"
```

## For Your Burp Suite Setup
Since you're working with the Docker Burp Suite setup, you can use the provided script:

```bash
# Navigate to the setup directory
cd space/idea/web_sec/docker-burp-setup

# Use the provided script to start Firefox with Burp proxy configuration
./setup/start-firefox-burp.sh
```

## Create an Alias (Optional)
To make it easier, you can add this to your `~/.zshrc` file:

```bash
# Add to ~/.zshrc
alias firefox="/Applications/Firefox.app/Contents/MacOS/firefox"

# Then reload your shell
source ~/.zshrc

# Now you can just use:
firefox
```

The most convenient method for regular use is `open -a Firefox`, but for your security testing with Burp Suite, use the provided script `./setup/start-firefox-burp.sh` which configures the proxy settings automatically.