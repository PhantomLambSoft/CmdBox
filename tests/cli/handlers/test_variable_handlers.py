import unittest
from unittest.mock import MagicMock, patch
import typer
from cmdbox.cli.handlers.variable_handlers import (
    AddVariableArgs,
    run_add_variable,
    run_get_variable,
    run_update_variable,
    run_list_variables,
    run_search_variables,
    run_delete_variable,
)


class TestVariableHandlers(unittest.TestCase):

    def setUp(self):
        self.mock_var_services = MagicMock()
        self.mock_tag_services = MagicMock()
        self.mock_console = MagicMock()
        self.get_var_services = lambda: self.mock_var_services
        self.get_tag_services = lambda: self.mock_tag_services
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_created")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_value")
    @patch("cmdbox.cli.handlers.variable_handlers.get_tags_interactive")
    def test_run_add_variable_interactive(
        self, mock_get_tags, mock_prompt_value, mock_prompt_name, mock_render
    ):
        args = AddVariableArgs(name=None, value=None, tags=None, interactive=True)
        mock_prompt_name.return_value = "var1"
        mock_prompt_value.return_value = "val1"
        mock_get_tags.return_value = ["tag1"]
        mock_var = MagicMock()
        self.mock_var_services.create_variable.return_value = mock_var
        mock_render.return_value = "rendered_created"

        run_add_variable(
            args=args,
            get_var_services=self.get_var_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_var_services.create_variable.assert_called_with(
            name="var1", value="val1", tags=["tag1"]
        )
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_created")
    def test_run_add_variable_non_interactive(self, mock_render):
        args = AddVariableArgs(
            name="var1", value="val1", tags=["tag1"], interactive=False
        )
        mock_var = MagicMock()
        self.mock_var_services.create_variable.return_value = mock_var
        mock_render.return_value = "rendered_created"

        run_add_variable(
            args=args,
            get_var_services=self.get_var_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_var_services.create_variable.assert_called_with(
            name="var1", value="val1", tags=["tag1"]
        )
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable")
    def test_run_get_variable(self, mock_render):
        mock_var = MagicMock()
        self.mock_var_services.get_variable.return_value = mock_var
        mock_render.return_value = "rendered_var"
        run_get_variable(
            name="var1",
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )
        self.mock_var_services.get_variable.assert_called_with("var1")
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_var")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable(self, mock_render):
        mock_var = MagicMock()
        mock_var.id = 1
        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var
        mock_render.return_value = "rendered_updated"

        run_update_variable(
            name="var1",
            value="new_val",
            new_name="new_name",
            set_pairs=None,
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_called_with(
            "var1", value="new_val", name="new_name"
        )
        self.mock_console.success.assert_called_with("Variable updated successfully.")
        mock_render.assert_called_once_with(mock_updated_var)
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_variable_no_fields(self):
        with self.assertRaises(typer.BadParameter):
            run_update_variable(
                name="var1",
                value=None,
                new_name=None,
                set_pairs=None,
                get_var_services=self.get_var_services,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_list_variables(self, mock_render):
        vars_ = ["v1", "v2"]
        self.mock_var_services.list_variables.return_value = vars_
        mock_render.return_value = "rendered_list"
        run_list_variables(
            limit=10,
            order_by="name",
            tags=["t1"],
            fields=["f1"],
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )
        self.mock_var_services.list_variables.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(vars_, title="Variables", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_search_variables(self, mock_render):
        vars_ = ["v1"]
        self.mock_var_services.search_variables.return_value = vars_
        mock_render.return_value = "rendered_search"
        run_search_variables(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )
        self.mock_var_services.search_variables.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        mock_render.assert_called_once_with(
            vars_, title="Search Results", fields=["f1"]
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_deleted")
    def test_run_delete_variable_success(self, mock_render):
        mock_var = MagicMock()
        self.mock_var_services.get_variable.return_value = mock_var
        self.mock_var_services.delete_variable.return_value = True
        mock_render.return_value = "rendered_deleted"

        run_delete_variable(
            name="var1",
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )

        self.mock_var_services.delete_variable.assert_called_with("var1")
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_deleted")

    def test_run_delete_variable_failure(self):
        self.mock_var_services.delete_variable.return_value = False
        run_delete_variable(
            name="var1",
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
