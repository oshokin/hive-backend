# Hive-Backend

Hive-Backend is an educational social network application.

## Installation

### Docker

1. Install Docker for your system.
    - [Docker Desktop for Mac](https://hub.docker.com/editions/community/docker-ce-desktop-mac/)
    - [Docker Desktop for Windows](https://hub.docker.com/editions/community/docker-ce-desktop-windows/)
    - [Docker for Linux](https://docs.docker.com/install/linux/docker-ce/ubuntu/)

2. Clone the repository:
    ```
    git clone https://github.com/your-username/hive-backend.git
    cd hive-backend
    ```

3. Build and run the application with Docker Compose:
    ```
    docker-compose up --build
    ```
   or use `make compose-up`.
   
4. The application should now be running on http://localhost:8080.

5. To clean up everything, run either of the following commands:
    ```
    docker-compose down -v --rmi all
    ```
   or use `make compose-clean`.

## API Endpoints

- **GET** `/ping`: Check if the API is alive.
- **GET** `/metrics`: Get Prometheus metrics about the API.

### Cities

- **GET** `/v1/city/list`: Get a list of all cities.

### User Randomizing Jobs

- **GET** `/v1/randomizing-job/list`: Get a list of all user randomizing jobs.
- **POST** `/v1/randomizing-job/create`: Create a new user randomizing job.
- **POST** `/v1/randomizing-job/cancel`: Cancel a user randomizing job.

### Users

- **POST** `/v1/user/create`: Create a new user.
- **POST** `/v1/user/login`: Authenticate a user and generate a JWT token.
- **POST** `/v1/user/logout`: Logout a user and invalidate the JWT token.
- **GET** `/v1/user/{id}`: Get a user by ID.
- **GET** `/v1/user/search`: Search for users.

## Postman Collection

The Postman collection for this project is located at `/postman/hive-backend.json`. You can import this file into Postman to test the endpoints.

To import the collection:

1. Open Postman.
2. Click on the "Import" button.
3. Click "Choose Files" and select `hive-backend.json`.
4. The collection will now be available in Postman.