"""Integration checks for the local read-only explorer proxy."""

import json
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import explorer_readonly_proxy as proxy


class StubRPC(BaseHTTPRequestHandler):
    seen = []

    def do_GET(self):
        self.seen.append(self.path)
        body = json.dumps({"result": {"path": self.path}}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_):
        pass


class ExplorerProxyTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.upstream = ThreadingHTTPServer(("127.0.0.1", 0), StubRPC)
        cls.proxy = ThreadingHTTPServer(("127.0.0.1", 0), proxy.ReadonlyExplorerHandler)
        cls.original_origin = proxy.RPC_ORIGIN
        proxy.RPC_ORIGIN = f"http://127.0.0.1:{cls.upstream.server_port}"
        cls.threads = [threading.Thread(target=server.serve_forever, daemon=True) for server in (cls.upstream, cls.proxy)]
        for thread in cls.threads:
            thread.start()

    @classmethod
    def tearDownClass(cls):
        proxy.RPC_ORIGIN = cls.original_origin
        for server in (cls.proxy, cls.upstream):
            server.shutdown()
            server.server_close()

    def setUp(self):
        StubRPC.seen.clear()
        self.base = f"http://127.0.0.1:{self.proxy.server_port}"

    def test_allows_only_status_and_specific_block_height(self):
        for path in ("/status", "/block?height=10", "/validators?height=10"):
            with urlopen(self.base + path) as response:
                self.assertEqual(json.load(response)["result"]["path"], path)
        self.assertEqual(StubRPC.seen, ["/status", "/block?height=10", "/validators?height=10"])

    def test_rejects_unsafe_methods_and_arbitrary_routes(self):
        for path in ("/broadcast_tx_commit", "/block?height=0", "/block?height=10&unsafe=1", "/status?foo=1"):
            with self.assertRaises(HTTPError) as error:
                urlopen(self.base + path)
            error.exception.close()
        with self.assertRaises(HTTPError) as error:
            urlopen(Request(self.base + "/status", data=b"payload", method="POST"))
        self.assertEqual(error.exception.code, 405)
        error.exception.close()
        self.assertEqual(StubRPC.seen, [])


if __name__ == "__main__":
    unittest.main()
