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

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_created")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    def test_run_add_tag_interactive(
        self, mock_prompt_desc, mock_prompt_name, mock_render
    ):
        args = AddTagArgs(name=None, description=None, interactive=True)
        mock_prompt_name.return_value = "tag1"
        mock_prompt_desc.return_value = "desc1"
        mock_tag = MagicMock()
        self.mock_tag_services.create_tag.return_value = mock_tag
        mock_render.return_value = "rendered_created"

        run_add_tag(
            args=args,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.create_tag.assert_called_with(
            name="tag1", description="desc1"
        )
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_created")
    def test_run_add_tag_non_interactive(self, mock_render):
        args = AddTagArgs(name="tag1", description="desc1", interactive=False)
        mock_tag = MagicMock()
        self.mock_tag_services.create_tag.return_value = mock_tag
        mock_render.return_value = "rendered_created"

        run_add_tag(
            args=args,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.create_tag.assert_called_with(
            name="tag1", description="desc1"
        )
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag")
    def test_run_get_tag(self, mock_render):
        mock_tag = MagicMock()
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_render.return_value = "rendered_tag"
        run_get_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_tag_services.get_tag.assert_called_with("tag1")
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_tag")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag(self, mock_render):
        mock_tag = MagicMock()
        mock_tag.id = 1
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag
        mock_render.return_value = "rendered_updated"

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
        self.mock_console.success.assert_called_with("Tag updated successfully.")
        mock_render.assert_called_once_with(mock_updated_tag)
        self.mock_console.print.assert_called_with("rendered_updated")

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

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_list_tags(self, mock_render):
        tags = ["t1", "t2"]
        self.mock_tag_services.list_tags.return_value = tags
        mock_render.return_value = "rendered_list"
        run_list_tags(
            limit=10,
            order_by="name",
            fields=["f1"],
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_tag_services.list_tags.assert_called_with(limit=10, order_by="name")
        mock_render.assert_called_once_with(tags, title="Tags", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_search_tags(self, mock_render):
        tags = ["t1"]
        self.mock_tag_services.search.return_value = tags
        mock_render.return_value = "rendered_search"
        run_search_tags(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_tag_services.search.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        mock_render.assert_called_once_with(tags, title="Search Results", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_deleted")
    def test_run_delete_tag_success(self, mock_render):
        mock_tag = MagicMock()
        self.mock_tag_services.get_tag.return_value = mock_tag
        self.mock_tag_services.delete_tag.return_value = True
        mock_render.return_value = "rendered_deleted"

        run_delete_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.delete_tag.assert_called_with("tag1")
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_deleted")

    def test_run_delete_tag_failure(self):
        self.mock_tag_services.delete_tag.return_value = False
        run_delete_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
