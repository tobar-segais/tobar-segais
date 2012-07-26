<%@ page import="org.tobarsegais.webapp.data.Index" %>
<%@ page import="org.tobarsegais.webapp.data.IndexChild" %>
<%@ page import="org.tobarsegais.webapp.data.IndexEntry" %>
<%@ page import="org.tobarsegais.webapp.data.IndexTopic" %>
<%@ page import="org.tobarsegais.webapp.data.Toc" %>
<%@ page import="org.tobarsegais.webapp.data.Topic" %>
<%@ page import="java.util.Iterator" %>
<%@ page import="java.util.Map" %>
<%@ page import="java.util.Stack" %>
<%--
  Created by IntelliJ IDEA.
  User: stephenc
  Date: 26/07/2012
  Time: 12:36
  To change this template use File | Settings | File Templates.
--%>
<%@ page contentType="text/html;charset=UTF-8" language="java" %>
<html>
<head>
    <title></title>
</head>
<body>
<%
    Map<String, Toc> tocs = (Map<String, Toc>) application.getAttribute("toc");
    Map<String, Index> contents = (Map<String, Index>) application.getAttribute("indices");
    Stack<Iterator<? extends IndexChild>> stack = new Stack<Iterator<? extends IndexChild>>();
    Iterator<String> iterator = contents.keySet().iterator();
    String key = iterator.next();
    key = iterator.next();

    Toc toc = tocs.get(key);
    Index index = contents.get(key);
    if (!index.getChildren().isEmpty()) {
%>
<ul><%
    stack.push(index.getChildren().iterator());
    IndexChild entry;
    int dept = 1;
    while (!stack.empty()) {
        Iterator<? extends IndexChild> cur = stack.pop();
        if (!cur.hasNext()) {
            if (dept > 1) {
                out.print("</ul></li>");
            }
            dept--;
        } else {
            entry = cur.next();
            stack.push(cur);
            if (entry instanceof IndexEntry) {
                IndexEntry indexEntry = (IndexEntry) entry;
                out.print("<li>" + indexEntry.getKeyword());
                if (indexEntry.getChildren() != null) {
                    if (!indexEntry.getChildren().isEmpty()) {
                        stack.push(indexEntry.getChildren().iterator());
                        dept++;
                        out.print("<ul>");
                    }
                } else {
                    out.print("</li>");
                }
            } else if (entry instanceof IndexTopic) {
                IndexTopic topicEntry = (IndexTopic) entry;
                String href = topicEntry.getHref();
                int i = href.indexOf('#');
                href = i == -1 ? href : href.substring(0, i);
                Topic topic = toc.lookupTopic(href);
%>
    <li><a href="<%=topicEntry.getHref()%>"><%=topicEntry.getTitle() == null ? (topic != null
            ? topic.getLabel()
            : "null") : topicEntry.getTitle()%>
    </a></li>
    <%
                }
            }
        }
    %>
</ul>
<%
    }
%>
</body>
</html>