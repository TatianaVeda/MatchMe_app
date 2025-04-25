#!/bin/bash

# Function to output an error message and terminate the script
function error_exit {
  echo "$1" >&2
  exit 1
}

npm install concurrently --save-dev

# Checking for Docker
if ! command -v docker &> /dev/null; then
  echo "Docker not found. Trying to install it..."
  # For Ubuntu/Debian:
  sudo apt-get update
  sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
  sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  sudo apt-get update
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io
  # Add the current user to the docker group
  sudo usermod -aG docker $USER
  # Start and enable the Docker service
  sudo systemctl start docker
  sudo systemctl enable docker
  # Check again, if it didn't install, terminate the script
  if ! command -v docker &> /dev/null; then
    error_exit "Failed to install Docker. Please install it manually and try again."
  fi
  echo "Docker successfully installed."
  
  # Checking if a session restart is needed
  if ! groups | grep -q docker; then
    echo "ATTENTION: You have been added to the docker group, but the changes have not taken effect."
    echo "To continue without restarting the session, execute the command: newgrp docker"
    echo "Or restart the user session (logout and log back in)."
    read -p "Do you want to execute the newgrp docker command now? (y/n): " choice
    if [ "$choice" = "y" ] || [ "$choice" = "Y" ]; then
      echo "Executing the newgrp docker command..."
      exec newgrp docker
    else
      echo "Continuing with the script. Some docker commands may require sudo."
    fi
  fi
else
  echo "Docker is already installed."
fi

# Checking for Docker Compose
if ! command -v docker-compose &> /dev/null; then
  echo "Docker Compose not found. Trying to install it..."
  # For Ubuntu/Debian:
  sudo apt-get update
  sudo apt-get install -y docker-compose
  # Check again, if it didn't install, terminate the script
  if ! command -v docker-compose &> /dev/null; then
    error_exit "Failed to install Docker Compose. Please install it manually and try again."
  fi
fi

echo "Starting PostgreSQL through Docker Compose..."
docker-compose up -d

# Checking if PostgreSQL is available
echo "Waiting 10 seconds for PostgreSQL to fully start..."
sleep 10
docker exec m_postgres pg_isready
if [ $? -ne 0 ]; then
  echo "PostgreSQL is not ready! Check the container logs:"
  docker logs m_postgres
  error_exit
fi
echo "PostgreSQL is running and ready to connect."

# Installing dependencies for backend
echo "Moving to the backend folder and installing Go dependencies..."
cd backend || { echo "Backend folder not found!"; error_exit; }

# If the go.mod file is missing, initialize the module
if [ ! -f "go.mod" ]; then
  echo "go.mod file not found. Initializing the module..."
  go mod init m/backend
fi

# Updating/creating the go.sum file through go mod tidy (calculates and writes checksums)
echo "Updating dependencies (go.sum)..."
go mod tidy


# Running the installation of additional dependencies (if your code processes the -deps flag)
echo "Running the installation of additional dependencies with 'go run main.go -deps'..."
go run main.go -deps

# Returning to the root directory of the project
cd ..

# Installing dependencies for frontend
echo "Moving to the frontend folder and installing Node.js dependencies..."
cd frontend || { echo "Frontend folder not found!"; error_exit; }

# Checking if npx (Create React App) is installed
if ! command -v npx &> /dev/null
then
    echo "Error: npx not found. Installing Node.js to continue."
    npm install
fi

echo "Installing dependent libraries..."
# Installing react-router-dom for routing and axios for HTTP requests.
npm install react-router-dom axios @mui/material @mui/icons-material @emotion/react @emotion/styled react-toastify
 
npm install formik yup @hookform/resolvers || error_exit "Error installing."


# Additional libraries for form validation (if needed, uncomment)
# npm install formik yup

# Returning to the root directory of the project
cd .. || error_exit "Failed to return to the root project."

echo "Environment successfully configured!"
echo "-------------------------------------------------"
echo "To start the application, use the command: npm run dev"
echo "-------------------------------------------------"

# After installing frontend dependencies
npm audit fix --force
