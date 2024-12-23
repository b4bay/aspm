# Start a new stage for the runtime image
FROM debian:bookworm-slim

# Install SQLite
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app/

# Copy the binary from the builder stage
COPY /aspmd /app/
ENV PATH="$PATH:/app"

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./aspmd"]