const regexNameLine = /([A-Z\s]+)\.\\r\\n+/g;

const Controller = {
  search: (ev) => {
    ev.preventDefault();
    Controller.updateTable('')
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    fetch(`/search?q=${data.query}`).then((response) => {
      response.text().then((result) => {
        // Highlights name line.
        const t = result.replaceAll(regexNameLine, `<strong>$1</strong><br />`);
        Controller.updateTable(JSON.parse(t));
      });
    });
  },

  updateTable: (results) => {
    const table = document.getElementById("table-body");
    table.innerHTML = "";

    for (let result of results) {
      const title = document.createElement("div");
      title.className = "title"
      title.innerText = result.title
      table.append(title);

      const text = document.createElement("pre");
      text.innerHTML = result.text;
      table.append(text);

      table.append(document.createElement("br"))
    }
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
