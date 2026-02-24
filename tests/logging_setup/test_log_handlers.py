import unittest
import logging
from cmdbox.logging_setup.log_handlers import SecretRedactionFilter


class TestSecretRedactionFilter(unittest.TestCase):

    def setUp(self):
        self.filter = SecretRedactionFilter()

    def test_redact_token(self):
        record = logging.LogRecord(
            "test",
            logging.INFO,
            "path",
            10,
            "User token=secret-token-123 session started",
            (),
            None,
        )
        self.filter.filter(record)
        self.assertEqual("User token=[REDACTED] session started", record.msg)

    def test_redact_password(self):
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "login password=my_password123", (), None
        )
        self.filter.filter(record)
        self.assertEqual("login password=[REDACTED]", record.msg)

    def test_case_insensitivity(self):
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "TOKEN=Secret PASSWORD=Private", (), None
        )
        self.filter.filter(record)
        self.assertEqual("TOKEN=[REDACTED] PASSWORD=[REDACTED]", record.msg)

    def test_spacing_around_equals(self):
        record = logging.LogRecord(
            "test",
            logging.INFO,
            "path",
            10,
            "token  =  secret1 password=  secret2 token=secret3",
            (),
            None,
        )
        self.filter.filter(record)
        self.assertEqual(
            "token  =  [REDACTED] password=  [REDACTED] token=[REDACTED]", record.msg
        )

    def test_multiple_redactions(self):
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "token=t1 and password=p1", (), None
        )
        self.filter.filter(record)
        self.assertEqual("token=[REDACTED] and password=[REDACTED]", record.msg)

    def test_non_string_message(self):
        # logging.LogRecord message can be an object
        msg_obj = {"key": "value"}
        record = logging.LogRecord("test", logging.INFO, "path", 10, msg_obj, (), None)
        self.filter.filter(record)
        self.assertEqual(msg_obj, record.msg)

    def test_none_message(self):
        record = logging.LogRecord("test", logging.INFO, "path", 10, None, (), None)
        self.filter.filter(record)
        self.assertIsNone(record.msg)

    def test_partial_match_prevention(self):
        # Should NOT redact if it's 'mytoken='
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "mytoken=123 bypassword=abc", (), None
        )
        self.filter.filter(record)
        self.assertEqual("mytoken=123 bypassword=abc", record.msg)

    def test_empty_value(self):
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "token=   password=value", (), None
        )
        self.filter.filter(record)
        # In 'token=   password=value', the 'password=value' is considered the value of 'token=' by \S+
        # because \S+ matches 'password=value' (it contains no spaces).
        self.assertEqual("token=   [REDACTED]", record.msg)

    def test_value_with_spaces_stops_redaction(self):
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "token=secret part2", (), None
        )
        self.filter.filter(record)
        # \S+ should only catch 'secret'
        self.assertEqual("token=[REDACTED] part2", record.msg)

    def test_filter_returns_true(self):
        record = logging.LogRecord(
            "test", logging.INFO, "path", 10, "message", (), None
        )
        result = self.filter.filter(record)
        self.assertTrue(result)
