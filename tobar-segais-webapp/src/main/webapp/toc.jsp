<%@ page import="org.tobarsegais.webapp.data.Entry" %>
<%@ page import="org.tobarsegais.webapp.data.Toc" %>
<%@ page import="java.util.Iterator" %>
<%@ page import="java.util.Map" %>
<%@ page import="java.util.Stack" %>
<%

    Map<String, Toc> contents = (Map<String, Toc>) application.getAttribute("toc");

%>
<!doctype html>
<html lang="en">
<head>
    <script src="js/jquery-latest.js"></script>
    <link rel="stylesheet" href="css/screen.css" type="text/css"/>
    <link rel="stylesheet" href="css/jquery.treeview.css" type="text/css"/>
    <script type="text/javascript" src="js/jquery.treeview.js"></script>
    <script>
        $(document).ready(function() {
            $("#toc").treeview({collapsed:true});
        });
    </script>

</head>
<body>
<ul id="toc" class="filetree">
    <%
        for (Map.Entry<String, Toc> bundleEntry : contents.entrySet()) {
            Entry entry = bundleEntry.getValue();
            String bundle = bundleEntry.getKey();
    %>
    <li>
        <%if (entry.getHref() != null) {%>
        <a target="content" href="/content/<%=bundle+"/"+entry.getHref()%>">
            <%}%>
        <span
                class="<%=entry.getChildren().isEmpty()?"file":"folder"%>"><%=entry.getLabel()%></span>
            <%if (entry.getHref() != null) {%>
        </a>
        <%}%>
        <%
            Stack<Iterator<? extends Entry>> stack = new Stack<Iterator<? extends Entry>>();
            if (!entry.getChildren().isEmpty()) {
                out.print("<ul>");
                stack.push(entry.getChildren().iterator());
                while (!stack.empty()) {
                    Iterator<? extends Entry> cur = stack.pop();
                    if (cur.hasNext()) {
                        entry = cur.next();
                        stack.push(cur);
                        out.print("<li>");
        %>
        <%if (entry.getHref() != null) {%>
        <a target="content" href="/content/<%=bundle+"/"+entry.getHref()%>">
            <%}%>
        <span
                class="<%=entry.getChildren().isEmpty()?"file":"folder"%>"><%=entry.getLabel()%></span>
            <%if (entry.getHref() != null) {%>
        </a>
        <%}%>
        <%
                        if (!entry.getChildren().isEmpty()) {
                            out.print("<ul>");
                            stack.push(entry.getChildren().iterator());
                        }
                    } else {
                        out.print("</ul></li>");
                    }
                }
            }
        %></li>
    <%
        }
    %>
</ul>
</body>
</html>
