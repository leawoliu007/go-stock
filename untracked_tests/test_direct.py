import socket
import ssl

def test_direct_ip_http():
    print("Testing HTTP port 80...")
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(5.0)
    try:
        sock.connect(('61.129.129.196', 80))
        req = (
            "GET /api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20 HTTP/1.1\r\n"
            "Host: push2.eastmoney.com\r\n"
            "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36\r\n"
            "Connection: close\r\n\r\n"
        )
        sock.sendall(req.encode())
        resp = b''
        while True:
            data = sock.recv(4096)
            if not data:
                break
            resp += data
        print("HTTP SUCCESS! Response length:", len(resp))
        print(resp[:500].decode(errors='ignore'))
    except Exception as e:
        print("HTTP FAILED:", e)

def test_direct_ip_https():
    print("Testing HTTPS port 443...")
    context = ssl.create_default_context()
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(5.0)
    try:
        sock.connect(('61.129.129.196', 443))
        conn = context.wrap_socket(sock, server_hostname='push2.eastmoney.com')
        req = (
            "GET /api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20 HTTP/1.1\r\n"
            "Host: push2.eastmoney.com\r\n"
            "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36\r\n"
            "Connection: close\r\n\r\n"
        )
        conn.sendall(req.encode())
        resp = b''
        while True:
            data = conn.recv(4096)
            if not data:
                break
            resp += data
        print("HTTPS SUCCESS! Response length:", len(resp))
        print(resp[:500].decode(errors='ignore'))
    except Exception as e:
        print("HTTPS FAILED:", e)

if __name__ == "__main__":
    test_direct_ip_http()
    print()
    test_direct_ip_https()
