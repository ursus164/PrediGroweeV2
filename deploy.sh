#!/bin/bash

# Exit on any error
set -e

echo "🚀 Starting deployment process..."

# Define base directory
BASE_DIR=~/app_v2/PrediGroweeV2

# Ensure we're in the right directory
cd $BASE_DIR

# Pull the latest changes from the master branch
echo "📥 Pulling latest changes from git..."
git pull origin master

# Load environment variables from .env.prod
echo "⚙️ Loading environment variables..."
set -a  # automatically export all variables
source .env.prod
set +a  # stop automatically exporting

# Create backup of current databases (optional but recommended)
#echo "💾 Creating database backups..."
#timestamp=$(date +%Y%m%d_%H%M%S)
#mkdir -p $BASE_DIR/db_backups/$timestamp

#services=("auth" "quiz" "stats" "images")
#for service in "${services[@]}"
#do
#    echo "Backing up ${service}_db..."
#    docker-compose -f docker-compose.prod.yml exec -T ${service}_db pg_dump -U ${service}_user ${service}_db > "$BASE_DIR/db_backups/$timestamp/${service}_db_backup.sql" || echo "Warning: Could not backup ${service}_db"
#done

# Start or update the services
echo "🔄 Starting/updating services..."
docker-compose -f docker-compose.prod.yml up  --build

# Wait for services to be healthy
#echo "🏥 Checking services health..."
#sleep 10

## Check if all services are running
#echo "📊 Service status:"
#docker-compose -f docker-compose.prod.yml ps
#
## Check for any containers that exited
#failed_containers=$(docker-compose -f docker-compose.prod.yml ps -q -a --filter "exited=1")
#if [ ! -z "$failed_containers" ]; then
#    echo "❌ Some containers failed to start. Checking logs..."
#    docker-compose -f docker-compose.prod.yml logs --tail=50
#    exit 1
#fi
#
#echo "✅ Deployment completed successfully!"
#echo "You can check the logs using: docker-compose -f docker-compose.prod.yml logs -f"