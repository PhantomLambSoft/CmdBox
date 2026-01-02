import unittest
from unittest.mock import MagicMock
from cmdbox.services.tag_services import TagServices
from cmdbox.models import Tag


class TestTagServices(unittest.TestCase):
    def setUp(self):
        self.mock_repo = MagicMock()
        self.services = TagServices(self.mock_repo)

    def test_create_tag(self):
        # Setup
        name = "test-tag"
        description = "A test tag"
        expected_tag = MagicMock(spec=Tag)
        self.mock_repo.create.return_value = expected_tag

        # Execute
        result = self.services.create_tag(name, description)

        # Assert
        self.assertEqual(result, expected_tag)
        self.mock_repo.create.assert_called_once_with(
            name=name, description=description
        )

    def test_update_tag(self):
        # Setup
        name = "test-tag"
        tag = MagicMock(spec=Tag)
        updated_tag = MagicMock(spec=Tag)
        self.mock_repo.get_by_name.return_value = tag
        self.mock_repo.update.return_value = updated_tag

        # Execute
        result = self.services.update_tag(name, description="new")

        # Assert
        self.assertEqual(result, updated_tag)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.update.assert_called_once_with(tag, description="new")

    def test_delete_tag(self):
        # Setup
        name = "test-tag"
        tag = MagicMock(spec=Tag)
        self.mock_repo.get_by_name.return_value = tag
        self.mock_repo.delete.return_value = True

        # Execute
        result = self.services.delete_tag(name)

        # Assert
        self.assertTrue(result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.delete.assert_called_once_with(tag)

    def test_get_tag(self):
        # Setup
        name = "test-tag"
        expected_tag = MagicMock(spec=Tag)
        self.mock_repo.get_by_name.return_value = expected_tag

        # Execute
        result = self.services.get_tag(name)

        # Assert
        self.assertEqual(result, expected_tag)
        self.mock_repo.get_by_name.assert_called_once_with(name)

    def test_get_tag_by_id(self):
        id_ = 123
        expected_tag = MagicMock(spec=Tag)
        self.mock_repo.get_by_id.return_value = expected_tag

        result = self.services.get_tag_by_id(id_)

        self.assertEqual(expected_tag, result)
        self.mock_repo.get_by_id.assert_called_once_with(id_)

    def test_list_tags(self):
        # Setup
        expected_tags = [MagicMock(spec=Tag), MagicMock(spec=Tag)]
        self.mock_repo.list_all.return_value = expected_tags

        # Execute
        result = self.services.list_tags(order_by="name", limit=10)

        # Assert
        self.assertEqual(result, expected_tags)
        self.mock_repo.list_all.assert_called_once_with("name", 10)

    def test_search(self):
        # Setup
        query = "test"
        fields = ["name"]
        expected_tags = [MagicMock(spec=Tag)]
        self.mock_repo.search.return_value = expected_tags

        # Execute
        result = self.services.search(query, fields, limit=10)

        # Assert
        self.assertEqual(result, expected_tags)
        self.mock_repo.search.assert_called_once_with(query, fields, 10)
