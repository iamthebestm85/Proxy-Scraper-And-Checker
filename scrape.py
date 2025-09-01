import asyncio
import aiohttp
from aiohttp import ClientSession
import os
from colorama import Fore, Style, init

init(autoreset=True)

async def fetch_proxies(session, url):
    try:
        async with session.get(url, timeout=10) as response:
            if response.status == 200:
                text = await response.text()
                proxies = [line.strip() for line in text.splitlines() if line.strip() and ':' in line]
                return proxies
            else:
                print(f"{Fore.YELLOW}Warning: Failed to fetch {url} (status: {response.status})")
                return []
    except Exception as e:
        print(f"{Fore.RED}Error fetching {url}: {str(e)}")
        return []

async def main():
    if not os.path.exists('links.txt'):
        print(f"{Fore.RED}Error: links.txt not found!")
        return

    with open('links.txt', 'r') as f:
        links = [line.strip() for line in f if line.strip()]

    if not links:
        print(f"{Fore.YELLOW}No links found in links.txt.")
        return

    print(f"{Fore.GREEN}Found {len(links)} links to process.")

    all_proxies = set()

    async with ClientSession() as session:
        tasks = [fetch_proxies(session, url) for url in links]
        results = await asyncio.gather(*tasks, return_exceptions=True)

        for result in results:
            if isinstance(result, list):
                all_proxies.update(result)
            elif isinstance(result, Exception):
                print(f"{Fore.RED}Exception in task: {str(result)}")

    with open('proxies.txt', 'w') as f:
        for proxy in sorted(all_proxies):
            f.write(f"{proxy}\n")

    print(f"{Fore.GREEN}Successfully saved {len(all_proxies)} unique proxies to proxies.txt.")

if __name__ == "__main__":
    asyncio.run(main())
