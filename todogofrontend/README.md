

Final Project Definition (locked)
Stack

Backend: Go (net/http)

Frontend: Next.js

Datastore: Postgres

Auth: Session-based (email + password or token)

Deployment: Single instance (local or cloud)

No microservices. No queues (yet). No caches.

Backend: Endpoints (final)

Keep it to three endpoints.

1. POST /login

Authenticates user

Creates a session

Writes to Postgres

This endpoint is intentionally write-heavy.

2. POST /todos

Authenticated

Creates a todo

Writes to Postgres

Another write-heavy path.

3. GET /todos

Authenticated

Fetches latest todos

Read-heavy path

This is the endpoint you will protect later.

Database Schema (simple, but intentional)
users

id

email

password_hash

created_at

sessions

id

user_id

expires_at

todos

id

user_id

title

completed

created_at

Indexes:

sessions.user_id

todos.user_id

That’s it.

Act I: Initial Architecture (before fix)

Single Go service.
Single shared DB connection pool.

What this means

Login spikes compete with todo reads

Writes compete with reads

No endpoint is protected

This is where the failure will emerge.

Load Model (very important)

Your load test will simulate realistic usage, not synthetic noise.

Traffic mix

40% POST /login

40% POST /todos

20% GET /todos

This is intentionally hostile to reads.

Expected Failure (what you’re looking for)

Under bursty traffic:

GET /todos p95 latency spikes

Errors appear on reads

CPU stays low

DB connections are exhausted

Goroutines increase

This gives you your lesson:

Reads failed not because they were slow, but because they were crowded out.

Act III: The Single Fix (locked)
Isolate DB pools

One pool for /login

One pool for /todos writes

One protected pool for /todos reads

All pointing to the same Postgres DB.

What changes after the fix

Reads stabilize

Writes may degrade

System becomes predictable

You didn’t add power.
You added boundaries.

What you will record (for posting)

Screen capture of UI working during load

Metrics screenshot (latency before/after)

Short clip of k6 running

This becomes your video attachment.

How this turns into a strong LinkedIn post

You’ll say:

“I stress-tested a simple todo app and the first thing to fail was not the database or the CPU.
It was my assumption that all endpoints deserved equal access to shared resources.”

That’s senior-level insight.

Your immediate next steps (do these in order)

Build the backend with one shared DB pool

Build the Next.js UI (simple)

Verify functionality at low load

Add basic metrics

Run burst load test

Capture failure

Apply pool isolation

Rerun test

