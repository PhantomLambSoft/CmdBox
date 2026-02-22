import unittest
from unittest.mock import patch
from cmdbox.cli.commands import tag_crud
from cmdbox.cli.handlers import tag_handlers


class TestTagCrud(unittest.TestCase):

    @patch("cmdbox.cli.commands.tag_crud.tag_handlers.run_add_tag")
    @patch("cmdbox.cli.commands.tag_crud.container")
    def test_add(self, mock_container, mock_run_add_tag):
        # Setup
        name = "test-tag"
        description = "test description"
        interactive = False

        # Execute
        tag_crud.add(name=name, description=description, interactive=interactive)

        # Verify
        mock_run_add_tag.assert_called_once()
        args = mock_run_add_tag.call_args[1]["args"]
        self.assertIsInstance(args, tag_handlers.AddTagArgs)
        self.assertEqual(args.name, name)
        self.assertEqual(args.description, description)
        self.assertEqual(args.interactive, interactive)

        self.assertEqual(
            mock_run_add_tag.call_args[1]["get_tag_services"],
            mock_container.get_tag_services,
        )
        self.assertEqual(
            mock_run_add_tag.call_args[1]["get_console"], mock_container.get_console
        )

    @patch("cmdbox.cli.commands.tag_crud.tag_handlers.run_get_tag")
    @patch("cmdbox.cli.commands.tag_crud.container")
    def test_get(self, mock_container, mock_run_get_tag):
        # Setup
        name = "test-tag"

        # Execute
        tag_crud.get(name=name)

        # Verify
        mock_run_get_tag.assert_called_once_with(
            name=name,
            get_tag_services=mock_container.get_tag_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.tag_crud.tag_handlers.run_update_tag")
    @patch("cmdbox.cli.commands.tag_crud.container")
    def test_update(self, mock_container, mock_run_update_tag):
        # Setup
        name = "test-tag"
        description = "new description"
        new_name = "new-name"
        set_pairs = ["key1=val1"]

        # Execute
        tag_crud.update(
            name=name,
            description=description,
            new_name=new_name,
            set_=set_pairs,
            edit_mode=False,
            edit_fields=None,
        )

        # Verify
        mock_run_update_tag.assert_called_once_with(
            name=name,
            description=description,
            new_name=new_name,
            set_pairs=set_pairs,
            edit_mode=False,
            edit_fields=None,
            get_tag_services=mock_container.get_tag_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.tag_crud.tag_handlers.run_list_tags")
    @patch("cmdbox.cli.commands.tag_crud.container")
    def test_list_tags(self, mock_container, mock_run_list_tags):
        # Setup
        order = "name"
        limit = 5
        fields = ["name", "description"]

        # Execute
        tag_crud.list_tags(order=order, limit=limit, fields=fields)

        # Verify
        mock_run_list_tags.assert_called_once_with(
            limit=limit,
            fields=fields,
            order_by=order,
            get_tag_services=mock_container.get_tag_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_tag_display_field_resolver,
        )

    @patch("cmdbox.cli.commands.tag_crud.tag_handlers.run_search_tags")
    @patch("cmdbox.cli.commands.tag_crud.container")
    def test_search(self, mock_container, mock_run_search_tags):
        # Setup
        term = "search-term"
        limit = 20
        search_fields = ["name"]
        fields = ["description"]

        # Execute
        tag_crud.search(
            term=term, limit=limit, search_fields=search_fields, fields=fields
        )

        # Verify
        mock_run_search_tags.assert_called_once_with(
            term=term,
            limit=limit,
            search_fields=search_fields,
            fields=fields,
            get_tag_services=mock_container.get_tag_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_tag_display_field_resolver,
            get_search_field_resolver=mock_container.get_tag_search_field_resolver,
        )

    @patch("cmdbox.cli.commands.tag_crud.tag_handlers.run_delete_tag")
    @patch("cmdbox.cli.commands.tag_crud.container")
    def test_delete(self, mock_container, mock_run_delete_tag):
        # Setup
        name = "test-tag"

        # Execute
        tag_crud.delete(name=name)

        # Verify
        mock_run_delete_tag.assert_called_once_with(
            name=name,
            get_tag_services=mock_container.get_tag_services,
            get_console=mock_container.get_console,
        )
