from aiogram.fsm.state import State, StatesGroup


class NeuroStates(StatesGroup):
    waiting_prompt = State()
