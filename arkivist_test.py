import arkivist as arkv
import pathlib as pl
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


@fixture
def books_folder(pytestconfig: pyt.Config) -> pl.Path:
    sample_books = pytestconfig.rootpath.joinpath("sample_books")
    samples_copy = pl.Path().joinpath(
        "tmp", f"sample_books_{dt.datetime.now().isoformat()}"
    )
    sh.copytree(sample_books, samples_copy)
    yield samples_copy
    sh.rmtree(samples_copy)


@fixture
def book_of_unkown_isbn(pytestconfig: pyt.Config):
    # TODO: add test covering this case
    return pytestconfig.rootpath.joinpath(
        "sample_books",
        "By-ISBN",
        "978-1-098-10483-2 ÷ Knowledge Graphs - Data in Context for Responsive Businesses ÷ O'Reilly",
        "book.pdf",
    )


@fixture
def book_of_known_isbn(pytestconfig: pyt.Config):
    # TODO: add test covering this case
    return pytestconfig.rootpath.joinpath(
        "sample_books",
        "By-Publisher",
        "O'Reilly",
        "Erlang Programming",
        "Erlang Programming - Francesco Cesarini.pdf",
    )


def books_test(books_folder: pl.Path):
    assert books_folder.exists()
    assert sum(1 for _ in arkv.books(books_folder)) == 4


def resolve_metadata_test(book_of_known_isbn: pl.Path):
    assert arkv.book_info(book_of_known_isbn) == arkv.BookInfo(
        isbn="9780596518189",
        title="Erlang Programming",
        year=2009,
        publisher="O'Reilly",
        extension="pdf",
    )
