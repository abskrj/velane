FROM oven/bun:1.1-alpine
WORKDIR /app
COPY package.json ./
RUN bun install
COPY runner.ts cgroup_limits.ts ./
EXPOSE 8080
CMD ["bun", "run", "runner.ts"]
