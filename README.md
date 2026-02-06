# Marian Apparitions

## TODO

- [ ] Better data
    - [ ] (Maybe have some CRUD app to do all this)
    - [ ] Remove useless Wikipedia-specific sections such as "References", "See also", "Press coverage", "Gallery"
    - [ ] Divide sections
    - [ ] Have each section be stored into another table, with: id, event_id, language_id, sorting, title, content
    - [x] Group similar section titles to have some coherence across all Events
    - [ ] Categorize the requests into things like: "prayer", "penance", "construct a sacred building"

- [x] display new data stored in event_blocks table
- [ ] validate that the /static/ routes are secure and that you can't access files outside of the static directory
- [ ] validate that the query parameters are sanitized to prevent XSS attacks

- [ ] Add flag for the country where the apparitions happened
- [x] Add a JOIN with marys_requests for display in the view.html
- [x] Make sort and filter coexist
    - [x] Make sure we can't add sort_by more than once
- [ ] Add a link to the wikipedia page for each apparition
- [ ] Add a map of the location of each apparition
- [ ] Add a link to the shrine for each apparition that has one built at the demand of Our Lady

- [ ] Add images to each apparition
- [ ] Add an excerpt for each apparition
- [ ] Add links to youtube videos for each apparition

