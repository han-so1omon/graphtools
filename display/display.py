import os
import sys
import websocket
import time
import json
try:
    import thread
except ImportError:
    import _thread as thread

from drawille import Canvas, line, polygon

c = Canvas()

def find_edge_coords(nodes, edge_nodevals):
    startx = -1; starty = -1; endx = -1; endy = -1
    for n in nodes:
        if n['value'] == edge_nodevals[0]:
            startx = n['coords']['x']
            starty = n['coords']['y']

        if n['value'] == edge_nodevals[1]:
            endx = n['coords']['x']
            endy = n['coords']['y']

    return startx, starty, endx, endy

def graph2canvas(graph_message):
    c.clear()
    graph_dict = json.loads(graph_message)

    # Draw graph to canvas by iterating through nodes
    for n in graph_dict['nodes']:
        # Draw node marker
        x = n['coords']['x']
        y = n['coords']['y']
        for xc,yc in polygon(center_x=x, center_y=y, sides=6, radius=3):
            c.set(xc, yc)

        # Draw edge line
        if n['edges']:
            for e in n['edges']:
                startx, starty, endx, endy = find_edge_coords(
                        graph_dict['nodes'],
                        e['nodevals'])
                for xc,yc in line(startx*2, starty*2, endx*2, endy*2):
                        c.set(xc, yc)

    print(c.frame())

def process_message(message):
    graph2canvas(message)

def on_message(ws, message):
    process_message(message)

def on_error(ws, error):
    print("error")
    print(error)

def on_close(ws):
    print("### end ###")

def on_open(ws):
    def run(*args):
        while True:
            time.sleep(1)
        time.sleep(1)
        ws.close()
        print("terminating websocket...")
    thread.start_new_thread(run, ())

if __name__ == "__main__":
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp(
            "ws://localhost:8900/ws",
            on_message = on_message,
            on_error = on_error,
            on_close = on_close)
    ws.on_open = on_open
    ws.run_forever()
