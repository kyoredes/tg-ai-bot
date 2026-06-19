from clients.schemas import ClientModel


async def get_profile_info_answer(client: ClientModel):
    tg_id = client.tg_id
    email = client.email

    answer = f"""
    👤Ваш профиль
    ```
    Телеграм id: `{tg_id}`
    ```
    """
    if email:
        answer += f"\nEmail: `{email}`"

    return answer
