# realTimeFileEditor

A real-time collaborative document editing backend service built with Go, WebSockets, and PostgreSQL.

## 🚧 Project Status

This project is under active development. Core WebSocket functionality and Docker support are implemented. Frontend (Next.js) and CRDT-based sync logic (e.g. Yjs) are not yet complete.

---

## 🧠 About the Project

This codebase powers a real-time collaborative document editor. It uses WebSockets to enable multiple users to edit a shared document simultaneously. It includes document session management, authentication placeholders, persistent storage via PostgreSQL, and PDF generation of documents.

**Current Stack:**

- Go (Gin, GORM)
- WebSocket (github.com/coder/websocket)
- PostgreSQL
- Docker & Docker Compose
- Planned: Yjs CRDT integration, Next.js UI

---

## 📁 Folder Structure

```
realTimeFileEditor/
├── assets/
│   └── fonts/
│       ├── Roboto-Bold.ttf
│       └── Roboto-Regular.ttf
├── cmd/
│   └── main.go
├── config/
│   ├── cloudinary.go
│   └── dbConfig.go
├── internal/
│   ├── controllers/
│   ├── handlers/
│   ├── jobs/
│   │   └── docCleanUp.go
│   ├── middlewares/
│   ├── model/
│   ├── repositories/
│   ├── router/
│   └── ws/
├── pkg/
│   ├── constants/
│   │   └── constant.go
│   ├── jwt/
│   │   └── session.go
│   └── utils/
│       ├── codeGenerator.go
│       ├── documentGenerator.go
│       ├── documentHandler.go
│       └── hashing.go
├── templates/
├── .env
├── .gitignore
├── ca.pem
├── dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## 🧪 Getting Started

### 1. Clone the Repo

```bash
git clone https://github.com/nobelium24/realTimeFileEditor.git
cd realTimeFileEditor
```

### 2. Set Up Environment Variables

Create a `.env` file with the following content:

```
DB_URI=postgres://user:password@localhost:5432/realtimedb?sslmode=disable
SSL_CERT_PATH=ca.pem
SALT=test24Ram@Inc
JWT_SECRET=u2hdu!d83@2u8&9ue8U*U*U98+8uU@WWI
SMTP_HOST=smtp.gmail.com
SMTP_PORT=465
SMTP_USER=ogunbaja24@gmail.com
SMTP_PASS=qqzn iacu ucpk uzlr 
CLOUD_NAME=woleogunba
API_KEY=473735162712444
API_SECRET=lB6yiWTohmHQDZ4YVQBHveB1TG8
FE_ROOT_URL=test

```

### 3. Run PostgreSQL

You must have PostgreSQL running locally with a database named `realtimedb`.
If needed, set it up manually or add a `docker-compose.yml` later to manage it.

### 4. Run the App (Dockerized Backend)

```bash
docker build -t realtime-editor .
docker run -p 8080:8080 --env-file .env realtime-editor
```

Or run locally with Go:

```bash
go run cmd/main.go
```

## 📡 WebSocket Testing

A `testScripts/` folder is available to help test the real-time collaboration functionality without a frontend.

| Script | What it does |
|--------|---------------|
| `generateToken.js` | Generates JWT tokens for testing secured endpoints. |
| `client.js` | Simulates a single client editing a document via WebSocket. |
| `multiEdit.js` | Simulates multiple users editing a document concurrently. |

To use:

```bash
cd testScripts
npm install
node generateToken.js


## 🛠️ Planned Features

- [ ] CRDT synchronization using Yjs
- [ ] Auth using JWT sessions
- [ ] Full collaborative frontend using Next.js
- [ ] Access control per document

---

## 🤔 Why This Project?

This project is an exploration of real-time collaboration technologies with a focus on WebSocket communication, system design, and data synchronization.

> I aim to build scalable real-time systems and better understand CRDTs, state management, and distributed collaboration.

---

## 📬 Contact

Built by [@nobelium24](https://github.com/nobelium24) — feel free to reach out!
