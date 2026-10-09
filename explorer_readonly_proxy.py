#!/usr/bin/env python3
"""Loopback-only CometBFT reader for a public block explorer via HTTPS Funnel.

Serve only the status, block, and validator queries. Keep CometBFT RPC on
127.0.0.1:26657; Tailscale Funnel should point at this reader, not CometBFT.
"""

import json
import re
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.error import URLError
from urllib.parse import parse_qs, urlsplit
from urllib.request import urlopen


RPC_ORIGIN = "http://127.0.0.1:26657"
ALLOWED = {"/status", "/block", "/validators"}
MAX_RESPONSE_BYTES = 4 * 1024 * 1024


class ReadonlyExplorerHandler(BaseHTTPRequestHandler):
    def respond(self, status, body):
        payload = json.dumps(body).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Cache-Control", "no-store")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_GET(self):
        parsed = urlsplit(self.path)
        if parsed.path not in ALLOWED:
            self.respond(404, {"error": "Unknown chain query"})
            return

        try:
            query = parse_qs(parsed.query, strict_parsing=True, keep_blank_values=True)
        except ValueError:
            self.respond(400, {"error": "Invalid query parameters"})
            return
        if parsed.path == "/status":
            if query:
                self.respond(400, {"error": "Unexpected query parameters"})
                return
            target = RPC_ORIGIN + "/status"
        else:
            height = query.get("height", [])
            if (
                len(query) != 1
                or len(height) != 1
                or not re.fullmatch(r"[1-9][0-9]{0,14}", height[0])
            ):
                self.respond(400, {"error": "A positive integer height is required"})
                return
            target = f"{RPC_ORIGIN}{parsed.path}?height={height[0]}"

        try:
            with urlopen(target, timeout=5) as response:
                data = response.read(MAX_RESPONSE_BYTES + 1)
            if len(data) > MAX_RESPONSE_BYTES:
                self.respond(502, {"error": "RPC response is too large"})
                return
            self.respond(200, json.loads(data))
        except (URLError, TimeoutError, OSError, ValueError):
            self.respond(502, {"error": "Testnet RPC is unavailable"})

    def do_POST(self):
        self.respond(405, {"error": "Read-only endpoint"})


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", 8765), ReadonlyExplorerHandler)
    print("Read-only Mythchain explorer proxy: http://127.0.0.1:8765", flush=True)
    server.serve_forever()
