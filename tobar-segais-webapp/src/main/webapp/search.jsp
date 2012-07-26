<%@ page import="org.apache.lucene.store.Directory" %>
<%@ page import="org.apache.lucene.search.Query" %>
<%@ page import="org.apache.lucene.queryParser.QueryParser" %>
<%@ page import="org.apache.lucene.util.Version" %>
<%@ page import="org.apache.lucene.analysis.standard.StandardAnalyzer" %>
<%@ page import="org.apache.lucene.index.IndexReader" %>
<%@ page import="org.apache.lucene.search.IndexSearcher" %>
<%@ page import="org.apache.lucene.search.TopScoreDocCollector" %>
<%@ page import="org.apache.lucene.search.ScoreDoc" %>
<%@ page import="org.apache.lucene.document.Document" %>
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
       String query = request.getParameter("query");
       if (query == null || query.isEmpty()) {
          %>
   <form action="search.jsp" method="GET">
       <input name="query">
       <input type="submit" value="submit">
   </form>
          <%
       } else {
           Directory index = (Directory) application.getAttribute("index");
           Query q = new QueryParser(Version.LUCENE_34, "contents", new StandardAnalyzer(Version.LUCENE_34)).parse(query);
           int hitsPerPage = 10;
               IndexReader reader = IndexReader.open(index);
               IndexSearcher searcher = new IndexSearcher(reader);
               TopScoreDocCollector collector = TopScoreDocCollector.create(hitsPerPage, true);
               searcher.search(q, collector);
               ScoreDoc[] hits = collector.topDocs().scoreDocs;

               // 4. display results
               out.println("Found " + hits.length + " hits.");
               for(int i=0;i<hits.length;++i) {
                 int docId = hits[i].doc;
                 Document d = searcher.doc(docId);
                 out.println((i + 1) + ". " + d.get("title"));
               }

               // searcher can only be closed when there
               // is no need to access the documents any more.
               searcher.close();
       }
   %>
</body>
</html>