import unittest
from unittest.mock import MagicMock, patch
import typer
from cmdbox.cli.handlers.tag_handlers import (
    AddTagArgs,
    run_add_tag,
    run_get_tag,
    run_update_tag,
    run_list_tags,
    run_search_tags,
    run_delete_tag,
)


class TestTagHandlers(unittest.TestCase):

    def setUp(self):
        self.mock_tag_services = MagicMock()
        self.mock_console = MagicMock()
        self.get_tag_services = lambda: self.mock_tag_services
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    def test_run_add_tag_interactive(self, mock_prompt_desc, mock_prompt_name):
        args = AddTagArgs(name=None, description=None, interactive=True)
        mock_prompt_name.return_value = "tag1"
        mock_prompt_desc.return_value = "desc1"
        self.mock_tag_services.add_tag.return_value = MagicMock()

        run_add_tag(
            args=args,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.add_tag.assert_called_with(
            name="tag1", description="desc1"
        )
        self.mock_console.success.assert_called()

    def test_run_add_tag_non_interactive(self):
        args = AddTagArgs(name="tag1", description="desc1", interactive=False)
        self.mock_tag_services.add_tag.return_value = MagicMock()

        run_add_tag(
            args=args,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.add_tag.assert_called_with(
            name="tag1", description="desc1"
        )

    def test_run_get_tag(self):
        self.mock_tag_services.get_tag.return_value = "fake_tag"
        run_get_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_tag_services.get_tag.assert_called_with("tag1")
        self.mock_console.print_tag.assert_called_with("fake_tag")

    def test_run_update_tag(self):
        mock_tag = MagicMock()
        mock_tag.id = 1
        self.mock_tag_services.get_tag.return_value = mock_tag
        self.mock_tag_services.get_tag_by_id.return_value = "updated_tag"

        run_update_tag(
            name="tag1",
            description="new_desc",
            new_name="new_name",
            set_pairs=[],
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_called_with(
            "tag1", description="new_desc", name="new_name"
        )
        self.mock_console.success.assert_called()
        self.mock_console.print_tag.assert_called_with("updated_tag")

    def test_run_update_tag_no_fields(self):
        with self.assertRaises(typer.BadParameter):
            run_update_tag(
                name="tag1",
                description=None,
                new_name=None,
                set_pairs=[],
                get_tag_services=self.get_tag_services,
                get_console=self.get_console,
            )

    def test_run_list_tags(self):
        self.mock_tag_services.list_tags.return_value = ["t1", "t2"]
        run_list_tags(
            limit=10,
            order_by="name",
            fields=["f1"],
            get_tag_services=self.get_tag_services,  # Note: tag_handlers.py uses get_var_services parameter name in run_list_tags but it's likely a typo in original code. I'll follow the code.
            get_console=self.get_console,
        )
        self.mock_tag_services.list_tags.assert_called_with(limit=10, order_by="name")
        self.mock_console.print_tag_list.assert_called_with(
            ["t1", "t2"], output_fields=["f1"]
        )

    def test_run_search_tags(self):
        self.mock_tag_services.search_tags.return_value = ["t1"]
        run_search_tags(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_tag_services.search_tags.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        self.mock_console.print_tag_list.assert_called_with(
            ["t1"], output_fields=["f1"]
        )

    def test_run_delete_tag_success(self):
        mock_tag = MagicMock()
        self.mock_tag_services.get_tag.return_value = mock_tag
        self.mock_tag_services.delete_tag.return_value = True

        run_delete_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.delete_tag.assert_called_with("tag1")
        self.mock_console.success.assert_called()
        self.mock_console.print_tag.assert_called_with(mock_tag)

    def test_run_delete_tag_failure(self):
        self.mock_tag_services.delete_tag.return_value = False
        run_delete_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
