FROM docker:27-cli
RUN apk add --no-cache python3
WORKDIR /validator
COPY deploy/validator/server.py ./server.py
COPY deploy/validator/project.py ./project.py
COPY deploy/validator/deployment.py ./deployment.py
COPY backend/internal/arena/grade_test.go ./grade_test.go
CMD ["python3", "-u", "server.py"]
