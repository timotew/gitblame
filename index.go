package main

import "net/http"

const indexHtmlContent = `<script src="https://code.jquery.com/jquery-3.3.1.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js"></script>
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/diff2html/2.11.3/diff2html.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/mustache.js/3.1.0/mustache.min.js"></script>

<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/diff2html/2.11.3/diff2html.min.css">
<style>
    @media (min-width: 768px) {
        .modal-xxl {
            width: 100%;
            max-width: 1200px;
        }
    }

    a {
        cursor: pointer;
    }
</style>

<body style="margin: 20px 20px 20px 20px">

<script id="settings-template" type="text/template">
    <nav class="navbar navbar-light bg-light">
        <span class="navbar-brand mb-0 h1">Commits for Project at : {{settings.path}}</span>
    </nav>
</script>

<script id="action-buttons-template" type="text/template">
    <div style="margin: 20px">
        <button id="btn-update" class="btn btn-warning" type="button" onclick="update()">
            {{#updating}}
            <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
            {{/updating}}
            Update Repository
        </button>
    </div>
</script>

<script id="dropdown-template" type="text/template">
    <div class="dropdown">
        <button class="btn btn-secondary dropdown-toggle" type="button" id="dropdownMenuButton" data-toggle="dropdown"
                aria-haspopup="true" aria-expanded="false" style="margin: 20px">
            Team list with number of commits
        </button>
        <div class="dropdown-menu" aria-labelledby="dropdownMenuButton">
            {{#authors}}
            <a class="dropdown-item" onclick="showCommits('{{index}}')">
                {{author.name}}&lt;{{author.email}}&gt;
                <span class="badge badge-primary badge-pill">{{commits.length}}</span>
            </a>
            {{/authors}}
        </div>
    </div>
</script>

<script id="list-group-template" type="text/template">
    {{#author}}
    <div class="alert alert-info" role="alert">
        User: {{author.name}}&lt;{{author.email}}&gt; has {{commits.length}} commit:
    </div>
    {{/author}}

    <div class="list-group">
        {{#commits}}
        <a class="list-group-item list-group-item-action flex-column align-items-start"
           onclick="showDiff('{{title}}','{{hash}}')">
            <div class="d-flex w-100 justify-content-between">
                <h6 class="mb-1">{{{title}}}</h6>
                <small>{{since}}</small>
            </div>
            <p class="mb-1">{{{body}}}</p>
        </a>
        {{/commits}}
    </div>
</script>

<script id="diff-dialog-template" type="text/template">
    <div class="modal fade" id="diffModalLong" tabindex="-1" role="dialog" aria-labelledby="diffModalLongTitle"
         aria-hidden="true">
        <div class="modal-dialog modal-xxl" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="diffModalLongTitle">{{{title}}}</h5>
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <div class="modal-body">
                    {{{body}}}
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                </div>
            </div>
        </div>
    </div>
</script>

<div class="settings"></div>
<div class="action-buttons"></div>
<div class="dropdown"></div>
<div class="list-group"></div>
<div class="diff-dialog"></div>

</body>

<script>
    function bind(elementId, json) {
        var template = $('#' + elementId + "-template").html();
        Mustache.parse(template);
        $("." + elementId).html(Mustache.render(template, json));
    }
</script>
<script>

    var authors;
    $(document).ready(function () {
        $.ajax({
            url: "/api/settings"
        }).done(function (data) {
            var settings = JSON.parse(data);
            bind("settings", {
                "settings": settings
            })
        });

        $.ajax({
            url: "/api"
        }).done(function (data) {
            authors = JSON.parse(data);
            authors.forEach(function (author, index) {
                author.index = index
            });
            bind("dropdown", {
                "authors": authors
            })
        });
    });
    // ---------------------------------------

    bind("action-buttons", {});

    function update() {
        toggleButton("btn-update", "updating", false);
        $.ajax({url: "/api/update"}).done(function () {
            toggleButton("btn-update", "updating", true);
        })
    }

    function toggleButton(btnId, modelVar, enable) {
        var obj = {};
        obj[modelVar] = !enable;
        bind("action-buttons", obj);
        $("#" + btnId).prop('disabled', !enable);
    }

    // ---------------------------------------

    function showCommits(index) {
        var author = authors[index].author;
        var commits = authors[index].commits.map(function (commit) {
            var messageArr = commit.message.split("\n");
            var title = messageArr[0];
            var body = messageArr.slice(1, messageArr.length).join('<br/>');
            return {
                "title": escapeQuotes(title),
                "body": escapeQuotes(body),
                "hash": commit.hash,
                "since": 'Since ' + timeSince(new Date(commit.when)) + ' ago'
            }
        });
        bind("list-group", {
            "commits": commits,
            "author": author
        })
    }

    // ---------------------------------------

    function showDiff(title, hash) {
        $.ajax({
            url: "/api/diff/" + hash
        }).done(function (data) {
            var diffHtml = Diff2Html.getPrettyHtml(
                data, {inputFormat: 'diff', showFiles: true, matching: 'lines', outputFormat: 'line-by-line'}
            );
            bind("diff-dialog", {
                "title": title,
                "body": diffHtml
            });
            $('#diffModalLong').modal('show');
        });
    }

</script>

<script>

    function escapeQuotes(str) {
        return str.replace(/'/g, '&apos;').replace(/"/g, '&quot;');
    }

    //https://stackoverflow.com/a/3177838/171950
    function timeSince(date) {
        var seconds = Math.floor((new Date() - date) / 1000);
        var interval = Math.floor(seconds / 31536000);

        if (interval > 1) {
            return interval + " years";
        }
        interval = Math.floor(seconds / 2592000);
        if (interval > 1) {
            return interval + " months";
        }
        interval = Math.floor(seconds / 86400);
        if (interval > 1) {
            return interval + " days";
        }
        interval = Math.floor(seconds / 3600);
        if (interval > 1) {
            return interval + " hours";
        }
        interval = Math.floor(seconds / 60);
        if (interval > 1) {
            return interval + " minutes";
        }
        return Math.floor(seconds) + " seconds";
    }

</script>
`

func Index(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "text/html")
	writer.Write([]byte(indexHtmlContent))
}
