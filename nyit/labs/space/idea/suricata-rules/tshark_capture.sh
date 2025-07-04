tshark -r lab44.pcapng \
  -Y 'http.request.uri matches "(<|>|script|onload|alert|svg|%3[Cc]|%3[Ee]|%22|%27|%3[Dd])" || http.user_agent matches "(<|>|script|onload|alert|svg|%3[Cc]|%3[Ee]|%22|%27|%3[Dd])"' \
  -T fields -e frame.number -e ip.src -e tcp.srcport -e ip.dst -e tcp.dstport -e http.request.method -e http.request.uri -e http.user_agent |
awk -F'\t' '
{
  print "Frame:       " $1
  print "Source:      " $2 ":" $3
  print "Destination: " $4 ":" $5
  print "Method:      " $6
  print "URL:         " $7
  print "User-Agent:  " $8
  print "-----------------------------"
}'