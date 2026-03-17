from dataclasses import dataclass
from enum import Enum


@dataclass(frozen=True)
class CommandRecord:
    """
    Represents an immutable record of a command with an alias and a template.

    This class is designed to store the alias and the corresponding template for a
    specific command in a frozen dataclass format. It serves as a simple
    data container with no additional logic or functionality.

    Attributes:
        alias (str): The alias or shorthand used to reference the command.
        template (str): The string template representing the full command.
    """

    alias: str
    template: str


@dataclass(frozen=True)
class VariableRecord:
    """
    Represents an immutable record of a variable with an name and a value.

    This class is designed to store the name and the corresponding value for a
    specific variable in a frozen dataclass format. It serves as a simple
    data container with no additional logic or functionality.

    Attributes:
        name (str): The name of the variable.
        value (str): The value associated with the variable.
    """

    name: str
    value: str


class RefKind(str, Enum):
    """
    Enumeration for representing different kinds of references.

    This class serves as an enumeration to distinguish between various types of
    references such as commands or variables. It is particularly useful in scenarios
    where categorization or identification of reference types is needed for processing
    or decision-making purposes.

    Attributes:
        COMMAND (str): Represents a reference type that corresponds to a command.
        VARIABLE (str): Represents a reference type that corresponds to a variable.
    """

    COMMAND = "command"
    VARIABLE = "variable"


@dataclass(frozen=True)
class TraceStep:
    """
    Represents a single step in a trace sequence.

    This class is used to encapsulate information about a step in a tracing
    process. It is immutable and defined as a dataclass with frozen attributes
    to ensure its contents remain unchanged once initialized.

    Attributes:
        kind (RefKind): The type or category of the trace step.
        key (str): The unique identifier or key associated with this step in the trace.
        expanded_to (str): The resulting value or destination that the key is expanded to.
    """

    kind: RefKind
    key: str
    expanded_to: str


@dataclass
class ResolveResult:
    """
    Represents the result of a resolution process.

    This class is designed to encapsulate the details of a resolution operation,
    including the resulting text output and the trace of steps taken during the
    process. It serves as a structured way to manage and access the output and
    execution trace of a resolution operation.

    Attributes:
        text (str): The resulting text from the resolution process.
        trace (list[TraceStep]): The list of steps conducted during the resolution
            process, represented as TraceStep objects.
    """

    text: str
    trace: list[TraceStep]
