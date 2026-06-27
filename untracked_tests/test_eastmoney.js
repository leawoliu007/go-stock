

async function test() {
    const url = "https://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20";
    const resp = await fetch(url, {
        headers: {
            "User-Agent": "Mozilla/5.0",
            "Referer": "https://quote.eastmoney.com"
        }
    });
    const text = await resp.text();
    console.log(text);
}
test();
