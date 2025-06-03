## E2-E2 Encryption: [Ephemeral-to-Ephemeral] - [End-to-End]

### Overview:
E2-E2 encryption is a layered security protocol designed to establish an ephemeral-to-ephemeral encrypted channel between a client and server. The process involves two stages of key generation to ensure high-level security and forward secrecy.

### Process:

1. **K1 (Initial Key Exchange)**:
    - **Client** and **Server** generate their own **ECDH Curve25519 Public/Private Key Pairs**.
    - These public keys are exchanged between the client and server, allowing the generation of a shared secret **K1** through the Diffie-Hellman process.
    - **K1** establishes a secure **End-to-End (E2E)** encrypted channel for communication.

2. **K2 (Ephemeral Key Generation)**:
    - Once the secure channel is established with **K1**, both the client and server use their knowledge of **K1** and the **Ephemeral API Key (EAPI)** to generate a new key, **K2**.
    - **K2** is derived by hashing or using a Key Derivation Function (KDF) on **K1** and the **API Key** (a pre-shared key known only to the client and server).

3. **K3 (Key Combination)**:
    - **K2** is then used to derive **K3**, a key that becomes the primary secret for all future communications.
    - **K3** is derived without additional key exchanges, ensuring that even if **K1** is compromised, **K3** remains secure.
    - **K3** is generated independently by both the client and server using a combination of **K1** and the **API Key**, with no need for further transmissions.

4. **Use of K3**:
    - **K3** becomes the **full secret key** for both the client and server, handling all encrypted communication going forward.
    - This key can also be used to dynamically generate **Ephemeral API Keys (EAPI)** for session-based operations or further authentication.
    - With **K3**, both the client and server maintain a secure, shared secret for all communication without the need for additional exchanges.

### Benefits:
- **Forward Secrecy**: Compromise of **K1** does not expose **K3**, as **K3** is derived using information that is not transmitted.
- **Ephemeral Security**: By deriving **K2** and **K3** from **K1** and the **API Key**, the protocol ensures that each communication session is independent and secure.
- **Dynamic API Keys**: **K3** can be used to generate session-based API keys, adding another layer of security for authenticated requests.

### Requirements
- Go **1.24.3** or higher

### Starting the Server
The server library is in `project/Server`, with an example command in `project/Server/cmd`.
```bash
go run ./project/Server/cmd
```
The server listens on port `8080`.

### Running the Client
The client library is in `project/Client`, with an example command in `project/Client/cmd`.
Run the client in a separate terminal after the server has started:
```bash
go run ./project/Client/cmd
```
The `project/sh/run.sh` script can also be used to start the server and repeatedly run the client.

### Session Files
`client_session.json` and `server_sessions.json` are created automatically to store session data. These files are ignored by git.
