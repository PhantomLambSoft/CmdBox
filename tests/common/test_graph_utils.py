import unittest

from cmdbox.common.graph_utils import find_cycle


class TestFindCycle(unittest.TestCase):

    def test_returns_none_for_single_node_without_dependencies(self):
        graph = {"A": []}

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertIsNone(result)

    def test_returns_none_for_acyclic_graph(self):
        graph = {
            "A": ["B", "C"],
            "B": ["D"],
            "C": ["E"],
            "D": [],
            "E": [],
        }

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertIsNone(result)

    def test_detects_simple_cycle(self):
        graph = {
            "A": ["B"],
            "B": ["C"],
            "C": ["A"],
        }

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertEqual(["A", "B", "C", "A"], result)

    def test_detects_self_loop(self):
        graph = {"A": ["A"]}

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertEqual(["A", "A"], result)

    def test_ignores_cycle_not_reachable_from_start(self):
        graph = {
            "A": ["B"],
            "B": [],
            "X": ["Y"],
            "Y": ["X"],
        }

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertIsNone(result)

    def test_detects_cycle_reachable_from_start_even_with_other_branch(self):
        graph = {
            "A": ["B", "X"],
            "B": ["C"],
            "C": [],
            "X": ["Y"],
            "Y": ["X"],
        }

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertEqual(["X", "Y", "X"], result)

    def test_returns_first_cycle_found_by_dependency_order(self):
        graph = {
            "A": ["B", "D"],
            "B": ["C"],
            "C": ["B"],
            "D": ["E"],
            "E": ["D"],
        }

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertEqual(["B", "C", "B"], result)

    def test_does_not_report_false_cycle_when_dependency_repeats_but_not_on_stack(self):
        graph = {
            "A": ["B", "C"],
            "B": ["D"],
            "C": ["D"],
            "D": [],
        }

        result = find_cycle("A", lambda node: graph.get(node, []))

        self.assertIsNone(result)

    def test_supports_generator_dependencies(self):
        graph = {
            "A": ["B"],
            "B": ["C"],
            "C": ["A"],
        }

        def get_dependencies(node: str):
            for dep in graph.get(node, []):
                yield dep

        result = find_cycle("A", get_dependencies)

        self.assertEqual(["A", "B", "C", "A"], result)
