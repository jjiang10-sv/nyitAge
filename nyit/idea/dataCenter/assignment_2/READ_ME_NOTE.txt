All the flow spaces enabled bidirectional communication. 

Sometimes there is latency in within Pink-flowspace as the blue slice overlaps. 

However, we have observed that running the pox controller with forwarding.l2_pairs instead of forwarding.l2_learning for pink and blue controller, made the communication work better. 
cmd eg., sudo ./pox.py forwarding.l2_pairs openflow.of_01 --address=127.0.0.1 --port=6000.