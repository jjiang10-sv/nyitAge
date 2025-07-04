#!/usr/bin/env bash

PCAPFILE=$1
LOG_LOCATION=/workspace/suricata-rules/logs/

if [ -z "$PCAPFILE" ] || [ ! -f "$PCAPFILE" ]; then
    echo "File ${PCAPFILE} doesnt seem to be there - please supply a pcap file."
    exit 1;
fi

if [ ! -d "$LOG_LOCATION" ]; then
    echo "Attempting to create Suricata log directory..."
    mkdir "$LOG_LOCATION"
else
    echo "Log location exists, removing previous content..."
    rm -rf "$LOG_LOCATION"/*
fi

# Run Suricata in offline mode (i.e. PCAP processing)
# suricata -c /etc/suricata/suricata.yaml -k none -r "$1" --runmode=autofp -l "$LOG_LOCATION"
suricata -c ./suricata.yaml -k none -r "$1" --runmode=autofp -l "$LOG_LOCATION"

#print out alerts
echo -e "\nAlerts:\n"
grep '"event_type":"alert"' "$LOG_LOCATION/eve.json" | jq '(.timestamp) + " | " + (.alert.gid|tostring) + ":" + (.alert.signature_id|tostring) + ":" + (.alert.rev|tostring) + " | " + (.alert.signature) + " | " + (.alert.category) + " | " + (.src_ip) + ":" + (.src_port|tostring) + " -> " + (.dest_ip) + ":" + (.dest_port|tostring)'

# If you have Evebox installed, you can uncomment this line to launch it in oneshot mode
# evebox oneshot "$LOG_LOCATION/eve.json"