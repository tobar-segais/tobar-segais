
<%@ tag import="org.apache.commons.lang3.StringEscapeUtils" %>
<%@ tag import="org.apache.lucene.store.Directory" %>
<%@ tag import="org.tobarsegais.webapp.ServletContextListenerImpl" %>
<%@ tag import="org.apache.lucene.queryParser.QueryParser" %>
<%@ tag import="org.apache.lucene.search.Query" %>
<%@ tag import="org.apache.lucene.index.IndexReader" %>
<%@ tag import="org.apache.lucene.search.IndexSearcher" %>
<%@ tag import="org.apache.lucene.search.TopScoreDocCollector" %>
<%@ tag import="org.apache.lucene.search.ScoreDoc" %>
<%@ tag import="java.text.MessageFormat" %>
<%@ tag import="java.net.URLEncoder" %>
<%@ tag import="org.apache.lucene.queryParser.ParseException" %>
<form class="form-search" method="get" action=".">
        <input name="query" type="search" class="input-large search-query"
               value="<%=request.getParameter("query")==null?"":StringEscapeUtils.escapeHtml4(request.getParameter("query"))%>"
                placeholder="Search">
        <button type="submit" class="btn"><i class="icon-search"></i></button>
    </form>
    <%
        String query = request.getParameter("query");
        if (query != null && !query.isEmpty()) {
            Directory index = ServletContextListenerImpl.getDirectory(application);
            QueryParser queryParser = ServletContextListenerImpl.getContentsQueryParser(application);
            try {
                Query q = queryParser.parse(query);
                int hitsPerPage = 200;
                IndexReader reader = IndexReader.open(index);
                IndexSearcher searcher = new IndexSearcher(reader);
                try {
                    TopScoreDocCollector collector = TopScoreDocCollector.create(hitsPerPage, true);
                    searcher.search(q, collector);
                    ScoreDoc[] hits = collector.topDocs().scoreDocs;
                    out.print("<span>");
                    out.print(MessageFormat.format("Found {0} hits:", hits.length));
                    out.print("</span>");
                    out.print("<ul>");
                    for (int i = 0; i < hits.length; ++i) {
                        int docId = hits[i].doc;
                        org.apache.lucene.document.Document d = searcher.doc(docId);
                        String href = d.get("href");
                        int hashIndex = href.indexOf('#');
                        String hash = hashIndex == -1 ? "" : href.substring(hashIndex);
                        href = hashIndex == -1 ? href : href.substring(0, hashIndex);
                        String url = href + "?query=" + URLEncoder.encode(query, "UTF-8") + hash;
                        out.print("<li><a href=\"");
                        out.print(request.getContextPath());
                        out.print("/docs/");
                        out.print(url);
                        out.print("\">");
                        out.print(StringEscapeUtils.escapeHtml4(d.get("title")));
                        out.print("</a></li>");
                    }
                    out.print("</ul>");
                } finally {
                    // searcher can only be closed when there
                    // is no need to access the documents any more.
                    searcher.close();
                }
            } catch (ParseException e) {
                out.print(StringEscapeUtils.escapeHtml4(e.getMessage()).replace("\n", "<br />"));
            }
        }
    %>
