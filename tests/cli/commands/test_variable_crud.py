import unittest
from unittest.mock import patch
from cmdbox.cli.commands import variable_crud
from cmdbox.cli.handlers import variable_handlers


class TestVariableCrud(unittest.TestCase):

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_add_variable")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_add(self, mock_container, mock_run_add_variable):
        # Setup
        name = "test-var"
        value = "test-value"
        tags = ["tag1", "tag2"]
        interactive = False

        # Execute
        variable_crud.add(name=name, value=value, tags=tags, interactive=interactive)

        # Verify
        mock_run_add_variable.assert_called_once()
        args = mock_run_add_variable.call_args[1]["args"]
        self.assertIsInstance(args, variable_handlers.AddVariableArgs)
        self.assertEqual(args.name, name)
        self.assertEqual(args.value, value)
        self.assertEqual(args.tags, tags)
        self.assertEqual(args.interactive, interactive)

        self.assertEqual(
            mock_run_add_variable.call_args[1]["get_var_services"],
            mock_container.get_variable_services,
        )
        self.assertEqual(
            mock_run_add_variable.call_args[1]["get_tag_services"],
            mock_container.get_tag_services,
        )
        self.assertEqual(
            mock_run_add_variable.call_args[1]["get_console"],
            mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_get_variable")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_get(self, mock_container, mock_run_get_variable):
        # Setup
        name = "test-var"

        # Execute
        variable_crud.get(name=name)

        # Verify
        mock_run_get_variable.assert_called_once_with(
            name=name,
            get_var_services=mock_container.get_variable_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_update_variable")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_update(self, mock_container, mock_run_update_variable):
        # Setup
        name = "test-var"
        value = "new value"
        new_name = "new-name"
        set_pairs = ["key1=val1"]

        # Execute
        variable_crud.update(
            name=name,
            value=value,
            new_name=new_name,
            set_=set_pairs,
            edit_mode=False,
            edit_fields="name,value",
        )

        # Verify
        mock_run_update_variable.assert_called_once_with(
            name=name,
            value=value,
            new_name=new_name,
            set_pairs=set_pairs,
            edit_mode=False,
            edit_fields="name,value",
            get_var_services=mock_container.get_variable_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_list_variables")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_list_vars(self, mock_container, mock_run_list_variables):
        # Setup
        order = "name"
        tags = ["tag1"]
        limit = 5
        fields = ["name", "value"]

        # Execute
        variable_crud.list_vars(order=order, tags=tags, limit=limit, fields=fields)

        # Verify
        mock_run_list_variables.assert_called_once_with(
            order_by=order,
            tags=tags,
            limit=limit,
            fields=fields,
            get_var_services=mock_container.get_variable_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_variable_display_field_resolver,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_search_variables")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_search(self, mock_container, mock_run_search_variables):
        # Setup
        term = "search-term"
        limit = 20
        search_fields = ["name"]
        fields = ["value"]

        # Execute
        variable_crud.search(
            term=term, limit=limit, search_fields=search_fields, fields=fields
        )

        # Verify
        mock_run_search_variables.assert_called_once_with(
            term=term,
            limit=limit,
            search_fields=search_fields,
            fields=fields,
            get_var_services=mock_container.get_variable_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_variable_display_field_resolver,
            get_search_field_resolver=mock_container.get_variable_search_field_resolver,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_delete_variable")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_delete(self, mock_container, mock_run_delete_variable):
        # Setup
        name = "test-var"

        # Execute
        variable_crud.delete(name=name)

        # Verify
        mock_run_delete_variable.assert_called_once_with(
            name=name,
            get_var_services=mock_container.get_variable_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_attach_tags")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_add_tags(self, mock_container, mock_run_attach_tags):
        # Setup
        name = "test-var"
        tags = ["tag1", "tag2"]

        # Execute
        variable_crud.add_tags(name=name, tags=tags)

        # Verify
        mock_run_attach_tags.assert_called_once_with(
            name=name,
            tag_names=tags,
            get_var_services=mock_container.get_variable_services,
            get_tag_services=mock_container.get_tag_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.variable_crud.variable_handlers.run_detach_tags")
    @patch("cmdbox.cli.commands.variable_crud.container")
    def test_remove_tags(self, mock_container, mock_run_detach_tags):
        # Setup
        name = "test-var"
        tags = ["tag1", "tag2"]

        # Execute
        variable_crud.remove_tags(name=name, tags=tags)

        # Verify
        mock_run_detach_tags.assert_called_once_with(
            name=name,
            tag_names=tags,
            get_var_services=mock_container.get_variable_services,
            get_tag_services=mock_container.get_tag_services,
            get_console=mock_container.get_console,
        )
