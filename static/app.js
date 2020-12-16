const regex = /([A-Z\s]+)\.\\r\\n+/g;

const Controller = {
  search: (ev) => {
    ev.preventDefault();
    Controller.updateTable('')
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.text().then((result) => {
        // Highlights name line.
        const t = result.replaceAll(regex, `<strong>$1</strong><br />`);
        Controller.updateTable(JSON.parse(t));
      });
    });
  },

  updateTable: (results) => {
    const table = document.getElementById("table-body");
    table.innerHTML = '';
    for (let result of results) {
      const d = document.createElement('pre');
      d.innerHTML = result;
      table.append(d);
      table.append(document.createElement('hr'));
    }
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
