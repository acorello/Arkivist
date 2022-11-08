import arkivist as arkv
from pathlib import Path, PurePath
import datetime as dt
import shutil as sh
import pytest as pyt
from pytest import fixture

"""markdown
# Goal 1

Given a folder with electronics books stored as `*.pdf` or `*.epub` files in various places

I want to organize them in the following paths:

- `$BASE/By-ISBN/{ISBN} • {Title} • {Year} • {Publisher}/{Title}.{ext}`
- `$BASE/By-Publisher/{Title} • {Year}/{Title}.{ext}`

## Challenge

1. Reliably extract book’s ISBN: the proxy for all other attributes.
    - Strategy 1: parse ePub or PDF in folder and search for the ISBN.
2. Fetch book’s attributes from online sources (is isbntools sufficient?)
"""

NO_ISBN_FOLDER = PurePath("No-ISBN")


@fixture
def books_folder(pytestconfig: pyt.Config) -> Path:
    sample_books = pytestconfig.rootpath.joinpath("sample_books")
    samples_copy = Path().joinpath(
        "tmp", f"sample_books_{dt.datetime.now().isoformat()}"
    )
    sh.copytree(sample_books, samples_copy)
    yield samples_copy
    sh.rmtree(samples_copy)


def sanity_check_test(books_folder: Path):
    assert books_folder.exists()
    assert sum(1 for _ in arkv.books(books_folder)) == 4


@fixture
def book_without_metadata(pytestconfig: pyt.Config):
    return pytestconfig.rootpath.joinpath(
        "sample_books",
        "By-ISBN",
        "978-1-098-10483-2 ÷ Knowledge Graphs - Data in Context for Responsive Businesses ÷ O'Reilly",
        "book.pdf",
    )


@pyt.mark.skip(reason="TODO")
def book_without_metadata_test(book_without_metadata: PurePath):
    ...


@fixture
def book_with_metadata(pytestconfig: pyt.Config):
    return pytestconfig.rootpath.joinpath(
        "sample_books",
        "By-Publisher",
        "O'Reilly",
        "Erlang Programming",
        "Erlang Programming - Francesco Cesarini.pdf",
    )


@fixture
def book_without_isbn():
    ...


@pyt.mark.skip(reason="TODO")
def books_without_isbn_test(book_without_isbn: PurePath):
    f"Books without ISBN are placed under {NO_ISBN_FOLDER}"
    ...


@pyt.mark.skip(reason="TODO")
def other_files_beside_books_test():
    ...


def resolve_metadata_test(book_with_metadata: Path):
    assert arkv.book_info(book_with_metadata) == arkv.BookInfo(
        isbn="9780596518189",
        title="Erlang Programming",
        year=2009,
        publisher="O'Reilly",
        extension="pdf",
    )
