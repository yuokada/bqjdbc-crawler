#!/usr/bin/env python3
import os
import sys
import zipfile

import requests
import structlog
from bs4 import BeautifulSoup

DRIVER_DOWNLOAD_URL = "https://cloud.google.com/bigquery/docs/reference/odbc-jdbc-drivers"
DRIVER_FILENAME = "GoogleBigQueryJDBC42.jar"

EXCLUDE_DRIVER_LIST = [
    "SimbaJDBCDriverforGoogleBigQuery42_1.5.4.1008.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.5.0.1001.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.3.3.1004.zip",
    "SimbaBigQueryJDBC42-1.3.2.1003.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.3.0.1001.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.25.1029.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.23.1027.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.22.1026.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.21.1025.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.19.1023.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.18.1022.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.16.1020.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.14.1017.zip",
    "SimbaJDBCDriverforGoogleBigQuery42_1.2.1.1001.zip",
    "SimbaJDBCDriverforGoogleBigQuery41_1.2.1.1001.zip",
]

# Get logger from structlog
logger = structlog.get_logger()


def fetch_page_content(url: str) -> str | None:
    response = requests.get(url)
    if response.status_code == 200:
        return response.text
    return None


def get_driver_download_links(page_content: str) -> list[str]:
    soup = BeautifulSoup(page_content, "html.parser")
    links = []
    for download_link in soup.find_all("a", href=True):
        if "jdbc" in download_link["href"]:
            links.append(download_link["href"])
    return links


def exclude_old_drivers(driver_links: list[str]) -> list[str]:
    # Extract the filename from the URL
    def extract_filename(l):
        return l.split("/")[-1]

    return [l for l in driver_links if extract_filename(l) not in EXCLUDE_DRIVER_LIST]


def download_jdbc_driver(driver_link: str, dest_dir: str = "downloads") -> bool:
    response = requests.get(driver_link)
    if response.status_code == 200:
        f = f"{dest_dir}/{driver_link.split('/')[-1]}"
        with open(f, "wb") as file:
            file.write(response.content)
        if zipfile.is_zipfile(f):
            return True
        else:
            logger.error(f"{f} is not a zip file.")
            return False
    return False


def extract_specific_jar(zip_path: str, extract_to: str = "downloads") -> bool:
    jar_name = DRIVER_FILENAME
    zip_file_name = ".".join(zip_path.split("/")[-1].split(".")[0:-1])

    with zipfile.ZipFile(zip_path, "r") as zip_ref:
        all_files = zip_ref.namelist()
        if jar_name not in all_files:
            logger.warn(f"{jar_name} not found in the zip file.")
            return False

        # Extract the specific jar file
        zip_ref.extract(jar_name, extract_to)
        os.rename(f"{extract_to}/{jar_name}", f"{extract_to}/{zip_file_name}-{jar_name}")
        logger.info(f"Extracted {jar_name} to {extract_to}")
        return True


def check_downloaded_file(driver_link: str) -> bool:
    with open("download_history.txt") as file:
        lines = [l.strip() for l in file.readlines()]
        if driver_link in lines:
            return True
    return False


def append_to_download_history(driver_link: str) -> None:
    if check_downloaded_file(driver_link):
        return

    with open("download_history.txt", "a") as file:
        file.write(driver_link + "\n")


if __name__ == "__main__":
    page_content = fetch_page_content(DRIVER_DOWNLOAD_URL)
    if page_content is None:
        logger.error("Failed to fetch the page content")
        sys.exit(1)

    jdbc_links = get_driver_download_links(page_content)
    for link in exclude_old_drivers(jdbc_links):
        if check_downloaded_file(link):
            logger.info(f"{link} has already been downloaded.")
            continue

        if link.endswith(".zip"):
            logger.debug(link)
            result = download_jdbc_driver(link)
            if not result:
                logger.error(f"Failed to download {link}")
                sys.exit(1)
            logger.info(f"Downloaded {link}")

            # Extract the specific jar file
            extract_result = extract_specific_jar(f"downloads/{link.split('/')[-1]}")
            if not extract_result:
                logger.error(f"Failed to extract the jar file from {link}")
                sys.exit(1)
            append_to_download_history(link)
            logger.info(f"Extracted {DRIVER_FILENAME} from {link}")
