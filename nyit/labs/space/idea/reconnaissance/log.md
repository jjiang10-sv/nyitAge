nmap -p 1-10000 -A -oN output.txt 172.21.0.3

tshark -i eth0 -f "tcp"