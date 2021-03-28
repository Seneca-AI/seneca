# Tha Rules

1. Follow the language's linting rules.
2. All functions shall have inline documentation which includes a description, typed parameters and typed return values.
    - Example:
            
            # generate_event_card generates an EventCard object for the given
            # time_period for the user with the given user_id
            #   Params:
            #       TimePeriod time_period: denotes the time period for which
            #                               we want to generate the event_card
            #       int user_id:            denotes the uuid of the user we 
            #                               want to generate the event_card object
            #   Returns:
            #       EventCard:              the event card
            def generate_event_card(time_period, user_id):
                ...
3. Branches shall be named with prefixes categorizing the work. For example:
    * feature/new_thing
    * cleanup/fix_thing
4. All custom exported non-primtive typed objects must implement an exported package interface, use protos otherwise. (Exporting objects typed from libraries is fine.)
    * exported: starts with a capital letter
    * custom: defined in the package
    * non-primitive: not string, int, etc