# Distributed File Storage

## 📌 Overview
This project is a **distributed file storage system** built to explore **distributed systems concepts**. It implements a **custom-built Peer-to-Peer (P2P) library** for efficient network communication and decentralized file management. The system supports **data replication** for fault tolerance and provides essential file operations such as **GET, READ, and WRITE**.

Currently, it is a work in progress, with planned enhancements and additional features being actively developed.

---

## 🚀 Features

### 🔹 Custom P2P Library
- Implements **peer-to-peer communication** for decentralized storage.
- Manages **peer discovery, network communication, and data consistency**.

### 🔹 Data Replication
- Ensures **data availability and fault tolerance** by replicating files across multiple nodes.
- Reduces the risk of **data loss due to node failures**.

### 🔹 Core File Operations
#### **1️⃣ GET Method**
- Retrieves **metadata or status information** about files in the distributed system.

#### **2️⃣ READ Method**
- Reads the **contents of a specific file** from the distributed storage.

#### **3️⃣ WRITE Method**
- Writes **data to a specific file** while ensuring **data consistency and replication**.

---

## Folder Structure

```plaintext
.
├── Makefile
├── README.md
├── bin
│   └── fs
├── go.mod
├── go.sum
├── main.go
├── p2p
│   ├── encoding.go
│   ├── handshake.go
│   ├── message.go
│   ├── tcp_transport.go
│   ├── tcp_transport_test.go
│   └── transport.go
├── package-lock.json
├── package.json
├── server.go
├── store.go
└── store_test.go
```

---

## 🔧 Future Enhancements
- **Sharding support** for improved scalability.
- **Consensus mechanism** for ensuring data integrity.
- **Dynamic node management** for auto-scaling and resilience.
- **Security improvements** including **encryption and authentication**.

---

🚀 **Distributed File Storage - A Peer-to-Peer Approach to Scalable Storage!**

