# Reddit clone

## About the Project

**rclone** is a web application designed to create a discussions topics. The application consists of two modules:
1. **frontend**: js application from https://github.com/d11z/asperitas
2. **backend**: a service responsible for data storage, validation and logic

The application provides seamless CRUD (Create, Read, Update, Delete) operations on data

---

## Technologies

- **golang**: net/http, gorilla/mux, uber/zap, jwt, pgx, mongo-driver libraries
- **PostgreSQL**: relational database for storing users data
- **MongoDB**: document database for storing posts
- **Docker**: to run containers with PostgreSQL, Mongo databases and with an application

---

## Installation and Setup

### Prerequisites

1. **golang 1.24.2 or higher**
2. **Docker**
3. **Make**

---

## Steps

1. **Clone the Repository**:
   ```bash  
   git clone https://github.com/teatah/rclone.git
   cd rclone
   ```
   
2. **Run project**:
   - Ensure that Docker is installed and running on your machine
   - Ensure that Make is installed
   - Run the following command to start:
   ```bash  
   make 
   ```
   
## Stop project

To stop project run:
```bash  
   make stop
   ```
