import datetime
import json
import os
import time
from threading import Thread
from typing import Dict, List

import pprint

pprinter = pprint.PrettyPrinter(indent=2)

DEFAULT_PERSONAL_SCHEMA_PATH = "data/personal_data/schema.json"
DEFAULT_PERSONAL_DATA_PATH = "data/personal_data/data.json"
OPTIONS_DATA_KEY = "options"
DEFAULT_SCHEMA = {
    "options":
        {
            "default": 0
        }
    }
DEFAULT_OPTION = "ROOT"

COMMAND_MAPPINGS = {
    "C": "confirm",
    "H": "help",
    "N": "no",
    "P": "print",
    "Y": "yes",
}

HELP_OPTIONS = {
    "[H]elp": "Print this help message.",
    "[P]rint": "Print your data",
    "Undo": "Undo your last recording."
}

class Record:

    def __init__(self, key: str, value: int, timestamp: datetime.date):
        self.key = key
        self.value = value
        self.timestamp = timestamp

    def __str__(self) -> str:
        return f"{self.key}: {self.value} at {self.timestamp}"

    def undo(self, user_data: Dict[str, any]):
        user_input = get_input(f"Undoing: {self.__str__}. Are you sure?", ["YES", "NO"])
        if user_input == "NO":
            return

        key_parts = self.key.split(".")

        sub_data = user_data
        for part in key_parts[:-1]:
            sub_data = user_data[part]

        sub_data[key_parts[len(key_parts) - 1]] -= self.value


def main():
    print("Welcome to activity_log!")
    print_help_message()

    minutes_str = input_with_prompt("How often, in minutes, would you like to receive reminders? Choose 0 for no reminders.")

    if not minutes_str.isdigit():
        print("Not a digit.")
        return

    reminder_minutes = int(minutes_str)

    user_schema = {}
    user_data = {}

    last_record = None

    while True:
        # Create data if it doesn't exist.
        if len(user_schema) == 0:
            user_schema = create_or_get_json_file("schema", DEFAULT_PERSONAL_SCHEMA_PATH)
        if len(user_data) == 0:
            user_data = create_or_get_json_file("data", DEFAULT_PERSONAL_DATA_PATH)

        if not dict_keys_equal(user_data, user_schema):
            user_input = get_input("There is a mismatch between your schema and data. Would you like to update the schema or the data?", ["schema", "data"])
            if user_input == "SCHEMA":
                with open(DEFAULT_PERSONAL_SCHEMA_PATH, "w") as f:
                    json.dump(data_to_schema(user_data), f)
            elif user_input == "DATA":
                user_input = get_input("Have you made a backup of your data?", ["YES", "NO"])
                if user_input == "no":
                    print_important("Well you better go do it.")
                    continue
                elif user_input == "yes":
                    pass

        last_record = record_answer(DEFAULT_OPTION, reminder_minutes, last_record, user_data[OPTIONS_DATA_KEY])

        # Update schema if data changed.
        if not dict_keys_equal(user_data, user_schema):
            with open(DEFAULT_PERSONAL_SCHEMA_PATH, "w") as f:
                json.dump(data_to_schema(user_data), f)

        with open(DEFAULT_PERSONAL_DATA_PATH, "w") as f:
                json.dump(user_data, f)

        print_important("success")


def print_important(msg: str):
    print("==================================================")
    print(msg)
    print("==================================================")

def input_with_prompt(question: str) -> any:
    user_input = input(f"{question} \n >>> ")
    if handle_buzzwords(user_input):
        return input_with_prompt(question)

    return user_input

def record_answer(
    parent: str,
    reminder_interval: datetime.timedelta,
    last_record: Record,
    user_data: Dict[any, str]
):
    level_options = dict_level_to_indices(user_data)
    for opt, val in level_options.items():
        print(f"{opt}: {val}")

    user_choice = input_with_prompt("Choose one of the options to record, or type a new one to add it to your schema.")
    if handle_buzzwords(user_choice, last_record, user_data):
        return last_record

    if user_choice == "":
        user_choice = "1"

    if user_choice.isdigit():
        user_choice = int(user_choice)
        if not user_choice in level_options:
            print_important("Not a valid option.")
            return record_answer(parent, reminder_interval, last_record, user_data)

        choice = level_options[user_choice]
        if isinstance(user_data[choice], dict):
            return record_answer(parent, reminder_interval, last_record, user_data[choice])

        user_minutes = input_with_prompt(f"How many minutes did you do this for? Last record: {last_record} -- or you may add a new sub-option.")

        if user_minutes.isdigit():
            minutes_to_add = int(user_minutes)
            user_data[choice] += minutes_to_add

            return Record(f"{parent}.{choice}", minutes_to_add, datetime.datetime.now())
        elif user_minutes == "":
            if reminder_interval == 0:
                print_important("Without a reminder interval, we cannot process empty input here.")
                return record_answer(parent, reminder_interval, last_record, user_data)
                
            if last_record is None:
                print_important("No last record to undo.")
                return record_answer(parent, reminder_interval, last_record, user_data)

            time_diff = datetime.datetime.now() - last_record.timestamp

            if time_diff > reminder_interval:
                time_diff = reminder_interval

            minutes_to_add = int(time_diff.total_seconds() / 60)
            user_data[choice] += minutes_to_add
            return Record(f"{parent}.{choice}", minutes_to_add, datetime.datetime.now())
        else:
            if user_choice == 1:
                print_important("Sorry, but you can never expand choice 1.")
                return record_answer(parent, reminder_interval, last_record, user_data)

            user_new_choice = user_minutes

            expand_field(choice, user_new_choice, user_data)
            return record_answer(parent, reminder_interval, last_record, user_data[choice])

    # User wants to add a new choice.
    else:
        expand_field("", user_choice, user_data)
        return record_answer(parent, reminder_interval, last_record, user_data)

def handle_buzzwords(input: str, last_record: Record, user_data: Dict[str, any]) -> bool:
    input = get_command(input)

    if input == "UNDO":
        if last_record is None:
            print_important("Cannot undo last record if last record is not known.")
            return True

        last_record.undo(user_data)
        last_record = None
        return True

    if input == "PRINT":
        pprinter.pprint(user_data)
        return True

    if input == "HELP":
        print_help_message()
        return True
    
def print_help_message():
    print("Change your schema.json file to manually edit your own options.")
    pprinter.pprint(HELP_OPTIONS)

def update_data_schema(schema_dict: Dict[str, any], data_dict: Dict[str, any]) -> Dict[str, any]:
    new_data_dict = {}
    for key, val in schema_dict.items():
        if isinstance(val, dict):
            if key not in data_dict:
                new_data_dict[key] = update_data_schema(val, {})
            else:
                new_data_dict[key] = update_data_schema(val, data_dict[key])
        else:
            if key not in data_dict:
                new_data_dict[key] = None
            else:
                new_value = data_dict[key]
                if isinstance(new_value, dict):
                    new_value = dict_sum(new_value)
                new_data_dict[key] = new_value
    return new_data_dict

def dict_sum(dictionary: Dict[str, any]) -> int:
    sum = 0
    for val in dictionary.values:
        if isinstance(val, dict):
            sum += dict_sum(val)
        elif isinstance(val, int):
            sum += val
        else:
            raise Exception("Dict has types besides subdicts and ints.")
    return sum

def data_to_schema(data_dict: Dict[str, any]) -> Dict[str, any]:
    schema_dict = {}
    for key, val in data_dict.items():
        if isinstance(val, dict):
            schema_dict[key] = data_to_schema(val)
        else:
            schema_dict[key] = None

    return schema_dict
        
def dict_keys_equal(lhs_dict: Dict[str, any], rhs_dict: Dict[str, any]) -> bool:
    for key, val in lhs_dict.items():
        if key not in rhs_dict:
            return False

        if isinstance(val, dict):
            return dict_keys_equal(val, rhs_dict[key])

    return True

def expand_field(old_field: str, new_field: str, user_data: Dict[any, str]):
    user_input = get_input("You are about to change your schema, are you sure?", ["YES", "NO"])
    if user_input == "NO":
        return

    if old_field == "":
        user_data[new_field] = 0
    else:
        user_data[old_field] = {
            old_field: user_data[old_field],
            new_field: 0
        }

def dict_level_to_indices(dictionary: Dict[str, any]) -> Dict[int, str]:
    indexes = {}
    counter = 1
    for key, _ in dictionary.items():
        indexes[counter] = key
        counter += 1

    return indexes

        
def get_input(question: str, allowed_responses: List[str]) -> str:
    user_input = input_with_prompt(f"{question}. Choose from: {allowed_responses}")
    while True:
        if user_input not in allowed_responses and get_command(user_input) not in allowed_responses:
            print_important(f"Input not allowed. Available options are {allowed_responses}")
            user_input = input_with_prompt(question)
            continue

        return user_input.upper()


def get_command(input: str) -> str:
    input = input.upper()
    if input in COMMAND_MAPPINGS:
        return COMMAND_MAPPINGS[input].upper()
    return input.upper()

def create_or_get_json_file(file_type: str, path: str) -> Dict[str, any]:
    if not os.path.exists(path):
        yes_or_no = get_input(
            f"""
                You don't seem to have a {file_type} file yet. Would you like to create a default file?
            """,
            ["YES", "NO"]
        )

        if yes_or_no == "N":
            raise Exception("No data file.")

        with open(path, "w") as f:
            json.dump(DEFAULT_SCHEMA, f)
            time.sleep(1)

    with open(path, "r") as f:
        return json.load(f)

def validate_schema(schema: Dict[str, any]):
    for key, value in schema.items():
        if not isinstance(key, str):
            raise Exception("All keys of the schema must be dictionaries")

        if isinstance(value, dict):
            validate_schema(value)
        elif value is not None:
            raise Exception("Values must be subdict or None.")


if __name__ == "__main__":
    main()
