# Default policies
iptables -P INPUT DROP
iptables -P OUTPUT DROP
iptables -P FORWARD DROP

# Allow pings to the router
iptables -A INPUT -i eth0 -p icmp --icmp-type echo-request -j ACCEPT
iptables -A OUTPUT -p icmp --icmp-type echo-reply -j ACCEPT

# Allow internal hosts to ping external hosts
iptables -A FORWARD -i eth1 -s 192.168.60.0/24 -o eth0 -p icmp --icmp-type echo-request -j ACCEPT
iptables -A FORWARD -i eth0 -d 192.168.60.0/24 -o eth1 -p icmp --icmp-type echo-reply -j ACCEPT

# Block external pings to internal hosts
iptables -A FORWARD -i eth0 -d 192.168.60.0/24 -p icmp --icmp-type echo-request -j DROP