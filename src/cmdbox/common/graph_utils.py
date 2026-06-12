from typing import Callable, Iterable


def find_cycle(
    start: str, get_dependencies: Callable[[str], Iterable[str]]
) -> list[str] | None:
    """
    Finds a cycle in a directed graph starting from a given node.

    The function detects cycles within a graph represented by a node and its
    dependencies. If a cycle is found, it returns the nodes forming the cycle
    in the order they are visited, ending with the node where the cycle starts
    again. If no cycle exists, the function returns None.

    Args:
        start: The starting node of the graph traversal. This is the initial
            node from which the cycle detection process begins.
        get_dependencies: A callable function that returns an iterable of
            dependent nodes (edges) for a given node. This function is used to
            traverse the graph.

    Returns:
        A list of nodes representing the cycle if a cycle is found, starting
        and ending with the repeating node. Returns None if no cycle is detected.
    """
    stack: list[str] = []
    on_stack: set[str] = set()
    visited: set[str] = set()

    def visit(node: str) -> list[str] | None:
        if node in on_stack:
            cycle_start = stack.index(node)
            return stack[cycle_start:] + [node]
        if node in visited:
            return None

        visited.add(node)
        stack.append(node)
        on_stack.add(node)

        for dep in get_dependencies(node):
            result = visit(dep)
            if result is not None:
                return result

        stack.pop()
        on_stack.discard(node)
        return None

    return visit(start)
