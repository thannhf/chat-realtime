cd sunny/backend/services/user-service/cmd 
cd sunny/backend/services/media-service/cmd
cd sunny/backend/services/chat-service/cmd
cd sunny/backend/services/api-gateway/cmd 
cd sunny/frontend
go run main.go
npm run dev

USER_SERVICE_URL="http://localhost:8001"
CHAT_SERVICE_URL="http://localhost:8081"
MEDIA_SERVICE_URL="http://localhost:8083"
API_GATEWAY="http://localhost:8080"