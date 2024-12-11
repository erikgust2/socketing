import struct
import socket
import sys

# Prepare a binary message: ID=42, Value=3.14, Flag='A'
messages = [ 
    struct.pack('<Idc', 42, 3.14, b'A'), 
    struct.pack('<Idc', 24, 2.71, b'B'),
    struct.pack('<Idc', 1, 0.99, b'C')
    ]

print(sys.getsizeof(messages[1]))
print(sys.getsizeof(42))
print(sys.getsizeof(0.99))
print(sys.getsizeof(b'A'))

# Debug: Print the raw binary data being sent
for i, message in enumerate(messages):
    print(f"Message {i + 1}: {message.hex()}")

# Connect to the Go server
with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
    s.connect(('localhost', 12345))
    for message in messages:
        s.sendall(message)