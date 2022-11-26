from dataclasses import dataclass

import isbnlib
import isbnlib.registry
from isbnlib import ISBNLibException

from .arkivist import ISBN, BookInfo

# Query datasources for book metadata


@dataclass
class NoMetadata:
    def __init__(self, isbn: ISBN):
        self.isbn = isbn


class GetBookInfoFailed:
    def __init__(self, isbn: ISBN, exceptions: list[Exception]):
        self.exceptions = exceptions


def get_book_info(isbn: ISBN) -> BookInfo | NoMetadata:
    # TODO: make this concurrent or parallel; does isbnlib already provides a method for that?
    # /Users/am/Projects/Arkivist/.venv/lib/python3.11/site-packages/isbnlib/registry.py
    exceptions = []
    for provider in isbnlib.registry.PROVIDERS:
        try:
            if metadata := isbnlib.meta(isbn.ean13, service=provider):
                return BookInfo(
                    isbn=metadata["ISBN-13"],
                    title=metadata["Title"].title(),
                    publisher=metadata["Publisher"].title(),
                    year=int(metadata["Year"]),
                )
        except ISBNLibException as e:
            exceptions.append(e)
            continue
    # if exceptions is empty we did not find any data for the ISBN
    if exceptions:
        return GetBookInfoFailed(isbn, exceptions)
    else:
        return NoMetadata(isbn)
