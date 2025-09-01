import asyncio
import aiohttp
from aiohttp import ClientSession, ClientTimeout
import os
import signal
import sys
from tqdm import tqdm
from colorama import Fore, Style, init
import threading

init(autoreset=True)

working_proxies = []
output_file = None
stop_event = threading.Event()

def signal_handler(sig, frame):
    stop_event.set()

async def check_proxy(session, proxy, pbar):
    proxy_url = f"http://{proxy}"
    test_url = "http://httpbin.org/ip"
    try:
        async with session.get(test_url, proxy=proxy_url, timeout=ClientTimeout(total=5)) as response:
            if response.status == 200:
                with open(output_file, 'a') as f:
                    f.write(f"{proxy}\n")
                working_proxies.append(proxy)
    except Exception:
        pass
    finally:
        pbar.update(1)
        if stop_event.is_set():
            return

async def main():
    global output_file

    signal.signal(signal.SIGINT, signal_handler)

    input_file = input(f"{Fore.CYAN}Enter the path to the input proxies file (e.g., proxies.txt): ").strip()
    if not os.path.exists(input_file):
        print(f"{Fore.RED}Error: Input file not found!")
        return

    output_file = input(f"{Fore.CYAN}Enter the path to the output file for working proxies (e.g., working_proxies.txt): ").strip()

    with open(input_file, 'r') as f:
        proxies = [line.strip() for line in f if line.strip() and ':' in line]

    if not proxies:
        print(f"{Fore.YELLOW}No proxies found in {input_file}.")
        return

    proxies = list(set(proxies))
    print(f"{Fore.GREEN}Loaded {len(proxies)} unique proxies to check.")

    with open(output_file, 'w') as f:
        pass

    semaphore = asyncio.Semaphore(200)

    async def bound_check(proxy, pbar):
        async with semaphore:
            if not stop_event.is_set():
                await check_proxy(session, proxy, pbar)

    async with ClientSession() as session:
        try:
            with tqdm(total=len(proxies), desc=f"{Fore.MAGENTA}Checking proxies", colour='GREEN', dynamic_ncols=True) as pbar:
                tasks = [bound_check(proxy, pbar) for proxy in proxies]
                await asyncio.gather(*tasks, return_exceptions=True)
        except asyncio.CancelledError:
            pass

    print(f"{Fore.GREEN}Finished checking. Total working proxies: {len(working_proxies)} saved to {output_file}.")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print(f"\n{Fore.YELLOW}Stopped by user. Total working proxies found: {len(working_proxies)} saved to {output_file} ")
        print(f"\n{Fore.YELLOW} THANKS FOR USING MY TOOL")
        print(f"\n{Fore.YELLOW} - m85.68 ")
        sys.exit(0)
