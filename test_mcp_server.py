#!/usr/bin/env python3
"""
Test script for Gemara MCP Server

This script tests the MCP server by sending JSON-RPC messages via stdio
and verifying responses. It can be used to manually test the server
or as part of automated testing.
"""

import json
import subprocess
import sys
import time
from typing import Dict, Any, Optional


class MCPServerTester:
    def __init__(self, server_path: str = "./gemara-mcp-server"):
        """Initialize the tester with the path to the MCP server binary."""
        self.server_path = server_path
        self.process: Optional[subprocess.Popen] = None
        self.request_id = 1

    def start_server(self):
        """Start the MCP server process."""
        print(f"Starting MCP server: {self.server_path}")
        self.process = subprocess.Popen(
            [self.server_path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=0
        )
        print("Server started")

    def stop_server(self):
        """Stop the MCP server process."""
        if self.process:
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
            print("Server stopped")

    def send_request(self, method: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Send a JSON-RPC request to the server."""
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method
        }
        if params:
            request["params"] = params

        self.request_id += 1
        request_json = json.dumps(request) + "\n"
        
        print(f"\n>>> Sending request: {method}")
        print(f"    {request_json.strip()}")
        
        if not self.process or not self.process.stdin:
            raise RuntimeError("Server not started")
        
        self.process.stdin.write(request_json)
        self.process.stdin.flush()

        # Read response
        response_line = self.process.stdout.readline()
        if not response_line:
            raise RuntimeError("No response from server")
        
        print(f"<<< Received response:")
        print(f"    {response_line.strip()}")
        
        try:
            response = json.loads(response_line.strip())
            return response
        except json.JSONDecodeError as e:
            print(f"Error parsing JSON: {e}")
            print(f"Raw response: {response_line}")
            raise

    def test_initialize(self):
        """Test the initialize method."""
        print("\n" + "="*60)
        print("TEST: Initialize")
        print("="*60)
        
        params = {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "test-client",
                "version": "1.0.0"
            }
        }
        
        response = self.send_request("initialize", params)
        
        if response.get("result"):
            print("✓ Initialize successful")
            print(f"  Server info: {response['result'].get('serverInfo', {})}")
            return True
        else:
            print(f"✗ Initialize failed: {response.get('error', 'Unknown error')}")
            return False

    def test_initialized(self):
        """Test the initialized notification."""
        print("\n" + "="*60)
        print("TEST: Initialized notification")
        print("="*60)
        
        response = self.send_request("notifications/initialized", {})
        print("✓ Initialized notification sent")
        return True

    def test_list_prompts(self):
        """Test listing available prompts."""
        print("\n" + "="*60)
        print("TEST: List Prompts")
        print("="*60)
        
        response = self.send_request("prompts/list")
        
        if response.get("result"):
            prompts = response["result"].get("prompts", [])
            print(f"✓ Found {len(prompts)} prompt(s)")
            for prompt in prompts:
                print(f"  - {prompt.get('name')}: {prompt.get('description', 'No description')}")
            return True
        else:
            print(f"✗ List prompts failed: {response.get('error', 'Unknown error')}")
            return False

    def test_get_prompt(self):
        """Test getting a specific prompt."""
        print("\n" + "="*60)
        print("TEST: Get Prompt")
        print("="*60)
        
        params = {
            "name": "gemara-system-prompt"
        }
        
        response = self.send_request("prompts/get", params)
        
        if response.get("result"):
            prompt = response["result"]
            print(f"✓ Prompt retrieved: {prompt.get('name')}")
            messages = prompt.get("messages", [])
            print(f"  Messages: {len(messages)}")
            for i, msg in enumerate(messages):
                role = msg.get("role", "unknown")
                content = msg.get("content", [])
                if content and len(content) > 0:
                    text = content[0].get("text", "")[:100]
                    print(f"    [{i+1}] {role}: {text}...")
            return True
        else:
            print(f"✗ Get prompt failed: {response.get('error', 'Unknown error')}")
            return False

    def run_all_tests(self):
        """Run all tests."""
        try:
            self.start_server()
            time.sleep(0.5)  # Give server time to start
            
            results = []
            
            # Initialize
            results.append(("Initialize", self.test_initialize()))
            
            # Send initialized notification
            results.append(("Initialized", self.test_initialized()))
            
            # List prompts
            results.append(("List Prompts", self.test_list_prompts()))
            
            # Get prompt
            results.append(("Get Prompt", self.test_get_prompt()))
            
            # Print summary
            print("\n" + "="*60)
            print("TEST SUMMARY")
            print("="*60)
            for test_name, passed in results:
                status = "✓ PASS" if passed else "✗ FAIL"
                print(f"{status}: {test_name}")
            
            total = len(results)
            passed = sum(1 for _, p in results if p)
            print(f"\nTotal: {passed}/{total} tests passed")
            
            return passed == total
            
        except Exception as e:
            print(f"\n✗ Test failed with error: {e}")
            import traceback
            traceback.print_exc()
            return False
        finally:
            self.stop_server()


def main():
    """Main entry point."""
    import argparse
    
    parser = argparse.ArgumentParser(description="Test Gemara MCP Server")
    parser.add_argument(
        "--server",
        default="./gemara-mcp-server",
        help="Path to the MCP server binary (default: ./gemara-mcp-server)"
    )
    parser.add_argument(
        "--build",
        action="store_true",
        help="Build the server before testing"
    )
    
    args = parser.parse_args()
    
    if args.build:
        print("Building server...")
        result = subprocess.run(["go", "build", "-o", "gemara-mcp-server", "./cmd/gemara-mcp-server"], 
                              capture_output=True, text=True)
        if result.returncode != 0:
            print(f"Build failed: {result.stderr}")
            sys.exit(1)
        print("Build successful")
    
    tester = MCPServerTester(args.server)
    success = tester.run_all_tests()
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
