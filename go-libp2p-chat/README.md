# libp2p Chat

Peer-to-peer chat app built using libp2p for Juno.

A decentralized chat system can be designed using a peer-to-peer (P2P) architecture, where each node in the network acts as both a client and a server. Here is a prototype design for such a chat system:

Each node in the network runs a P2P client and server. The client allows the node to connect to other nodes and the server allows other nodes to connect to it.

A Distributed Hash Table (DHT) is used for peer discovery and routing. Each node stores its own IP address and listening port in the DHT, along with a unique identifier. Other nodes can use the DHT to lookup and connect to a specific node based on its identifier.

When a node wants to send a message to another node, it first looks up the destination node's IP address and listening port in the DHT. It then establishes a direct connection to the destination node and sends the message.

If a node wants to broadcast a message to multiple nodes, it can use a flooding algorithm to send the message to all of its directly connected peers, who then forward the message to their directly connected peers, and so on.

To handle offline messaging, nodes can store messages in the DHT for a specific period of time, indexed by the destination node's identifier. When the destination node comes online, it can retrieve any stored messages from the DHT.

To secure the chat system, end-to-end encryption can be used to encrypt messages between nodes. Public-key cryptography can be used to establish secure connections and exchange encryption keys.

A user interface can be implemented using a web-based or mobile client, which communicates with the P2P client running on the user's device.


![img1](https://user-images.githubusercontent.com/3880512/187168129-21af280d-9718-41b5-b324-84dc4599f157.JPG)


![img3](https://user-images.githubusercontent.com/3880512/187168840-ba948e44-1e3b-4eb7-9ad5-62a1380cc524.png)

This is how the network looks  like after a node joins the Kademlia network. We're the yellow node on the lower right. Yellow nodes are participants of a chatroom.



![img4](https://user-images.githubusercontent.com/3880512/187168929-bfffd24e-65ce-4157-89f8-c053e6b57ccc.png)


This is how the network looks like after the node requests a chatroom. All the nodes in the chatroom are now connected directly and separately from the Kademlia network.





![Personalised p2p CHat](https://user-images.githubusercontent.com/3880512/188261250-288229b8-141d-4df3-a907-2a77ceb05715.gif)

Feature Set - 

1)GRPC based Chat server / client

2)DDOS/Malicious Attack Prevention Algorithm 

3)Content Moderation

Negative Categories 

a) Spam

b) Abuse

c) Adult content


Positive Categories 

a)Highly relevant – quality content

b)Less relevant

c)Fun messages

4)Fancy user handles

5)Topic based chat room

6)Messaging cooling period Algorithm

7)Refreshing participants Algorithm 

8)User block facility 

9)User Message history and P2P chat session storage and recovery/resumption in peers

10)Storing User semantics for P2P CHAT sessions to be plugged in analytics engine

