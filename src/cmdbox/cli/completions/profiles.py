def complete_profile_names(incomplete: str) -> list[str]:
    from cmdbox.container import get_profile_service

    service = get_profile_service()
    if incomplete == "":
        profs = service.list_profiles(order_by="last_updated", limit=20)
    else:
        profs = service.search(incomplete, fields="name", limit=20)
    return [x.name for x in profs]