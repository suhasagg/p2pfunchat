# libp2p Chat

Peer-to-peer chat app built using libp2p for Juno.


![img1](https://user-images.githubusercontent.com/3880512/187168129-21af280d-9718-41b5-b324-84dc4599f157.JPG)





![img3](https://user-images.githubusercontent.com/3880512/187168840-ba948e44-1e3b-4eb7-9ad5-62a1380cc524.png)

This is how the network looks  like after a node joins the Kademlia network. We're the yellow node on the lower right. Yellow nodes are participants of a chatroom.



![img4](https://user-images.githubusercontent.com/3880512/187168929-bfffd24e-65ce-4157-89f8-c053e6b57ccc.png)


This is how the network looks like after the node requests a chatroom. All the nodes in the chatroom are now connected directly and separately from the Kademlia network.


Start content analytics python module based on sentence transformers in start to create ML analytics tracker used by p2p chat.


Feature Set - 

1)GRPC based Chat server / client

2)DDOS/Malicious Attack Prevention Algorithm 

3)Content Moderation

Negative Categories - These messages are blocked and not reachable to other users.

a) Spam

b) Abuse

c) Adult content


Positive Categories 

a)Highly relevant – quality content for assigning badges, points to user to earn rewards

b)Less relevant

c)Fun messages

4)Fancy user handles

5)Topic based chat room

6)Messaging cooling period Algorithm

7)Refreshing participants Algorithm 

8)User block facility 

9)User Message history and P2P chat session storage and recovery/resumption in peers. 
  Batch mode message history sync to p2p chat database tracker for analytics and storage.
  Hbase based message database tracker server - 
  https://github.com/suhasagg/hbase-messageserver-as-p2pchat-database-tracker
  

10)Storing User semantics for P2P CHAT sessions to be plugged in analytics engine
   User chat semantics are available in bolt db so that other apps can utilise it for personalisation.


