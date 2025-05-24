#!/bin/bash

# Function to output error message and exit script
function error_exit {
  echo "$1" >&2
  exit 1
}

npm install concurrently --save-dev

# Check for Docker Compose
if ! command -v docker-compose &> /dev/null; then
  echo "Docker Compose not found. Trying to install it..."
  # For Ubuntu/Debian:
  sudo apt-get update
  sudo apt-get install -y docker-compose
  # Check again, if not installed, exit script
  if ! command -v docker-compose &> /dev/null; then
    error_exit "Failed to install Docker Compose. Please install it manually and try again."
  fi
fi


echo "Starting PostgreSQL via Docker Compose..."
docker-compose up -d

# Checking if PostgreSQL is ready
echo "Waiting 10 seconds for PostgreSQL to fully start..."
sleep 10
docker exec m_postgres pg_isready
if [ $? -ne 0 ]; then
  echo "PostgreSQL is not ready! Check container logs:"
  docker logs m_postgres
  error_exit
fi
echo "PostgreSQL is running and ready for connection."

# Moving to backend folder and installing Go dependencies...
echo "Moving to backend folder and installing Go dependencies..."
cd backend || { echo "Backend folder not found!"; error_exit; }

# If go.mod file is missing, initialize module
if [ ! -f "go.mod" ]; then
  echo "go.mod file not found. Initializing module..."
  go mod init m/backend
fi

# Updating/creating dependencies (go.sum) through go mod tidy...
echo "Updating dependencies (go.sum)..."
go mod sum

# Running dependency installation with 'go run main.go -deps'...
echo "Running dependency installation with 'go run main.go -deps'..."
go run main.go -deps

# Returning to project root directory
cd ..

# Moving to frontend folder and installing Node.js dependencies...
echo "Moving to frontend folder and installing Node.js dependencies..."
cd frontend || { echo "Frontend folder not found!"; error_exit; }

# Check if npx is installed (Create React App)
if ! command -v npx &> /dev/null
then
    echo "Error: npx not found. Installing Node.js to continue."
    npm install
fi

echo "Installing dependent libraries..."
# Installing react-router-dom for routing and axios for HTTP requests.
npm install react-router-dom axios @mui/material @mui/icons-material @emotion/react @emotion/styled react-toastify || error_exit "Failed to install dependencies."
npm install formik yup @hookform/resolvers || error_exit "Installation error."


# Additional libraries for form validation (can be uncommented if needed)
# npm install formik yup

# Returning to project root directory
cd .. || error_exit "Failed to return to project root directory."

echo "Environment successfully configured!"
echo "-------------------------------------------------"
echo "To run the application use command: npm run dev"
echo "-------------------------------------------------"
