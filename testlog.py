import argparse
from ast import parse
import asyncio
import websockets
import json
from uuid import uuid4

EVENT_TYPE = "PlayerMessage"

async def mineproxy(websocket, _):
    print('Connected')

    # Tell Minecraft to send all chat messages. Required once after Minecraft starts
    await websocket.send(
        json.dumps({
            "header": {
                "version": 1,                     # We're using the version 1 message protocol
                "requestId": str(uuid4()),        # A unique ID for the request
                "messageType": "commandRequest",  # This is a request ...
                "messagePurpose": "subscribe"     # ... to subscribe to ...
            },
            "body": {
                "eventName": EVENT_TYPE
            },
        }))

    try:
        # When MineCraft sends a message (e.g. on player chat), print it.
        async for msg in websocket:
            msg = json.loads(msg)
            print(json.dumps(msg, indent=2))
    except websockets.exceptions.ConnectionClosedError:
        print('Disconnected from MineCraft')

def main():
    global EVENT_TYPE
    parser = argparse.ArgumentParser("Minecraft websocket subscriber")
    parser.add_argument("--host", default="localhost", help="host to listen on, default localhost")
    parser.add_argument("--port", default=8000, help="port to listen on, default 8000")
    parser.add_argument("--event", default=EVENT_TYPE, help="event type to subscript to, defauly 'PlayerMessage'")
    args = parser.parse_args()
    EVENT_TYPE = args.event

    start_server = websockets.serve(mineproxy, host=args.host, port=args.port)
    print(f'Ready. On MineCraft, type /connect {args.host}:{args.port}')

    asyncio.get_event_loop().run_until_complete(start_server)
    asyncio.get_event_loop().run_forever()

if __name__ == "__main__":
    main()
