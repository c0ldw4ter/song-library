# Image of Nginx(Используем официальный образ Nginx)
FROM nginx:alpine

# Copy files from front-end to Nginx in contaner(Копируем файлы front-end в директорию Nginx)
COPY . /usr/share/nginx/html

# Open 80 port (Открываем порт 80 для HTTP)
EXPOSE 80

# Run Nginx(Запускаем Nginx)
CMD ["nginx", "-g", "daemon off;"]