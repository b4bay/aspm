# Create a minimal image with the Go binary and static files
FROM debian:bookworm-slim

# Set the Current Working Directory inside the container
WORKDIR /app

# Install necessary dependencies (e.g., for serving static files, etc.)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy backend files into container
COPY /aspm /app/
ENV PATH="$PATH:/app"

# To run with arguments
ENTRYPOINT []
CMD ["aspm"]
