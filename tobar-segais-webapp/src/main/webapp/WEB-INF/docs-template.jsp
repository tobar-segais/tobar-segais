<!DOCTYPE html>
<%--
  ~ Copyright 2012 Stephen Connolly
  ~
  ~ Licensed under the Apache License, Version 2.0 (the "License");
  ~ you may not use this file except in compliance with the License.
  ~ You may obtain a copy of the License at
  ~
  ~     http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing, software
  ~ distributed under the License is distributed on an "AS IS" BASIS,
  ~ WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  ~ See the License for the specific language governing permissions and
  ~ limitations under the License.
  --%>

<%@ page import="org.apache.commons.io.IOUtils" %>
<%@ page import="org.jsoup.Jsoup" %>
<%@ page import="org.jsoup.nodes.Document" %>
<%@ page import="org.tobarsegais.webapp.ServletContextListenerImpl" %>
<%@ page import="org.tobarsegais.webapp.data.TocEntry" %>
<%@ page import="org.tobarsegais.webapp.data.Toc" %>
<%@ page import="java.io.InputStream" %>
<%@ page import="java.net.JarURLConnection" %>
<%@ page import="java.net.URL" %>
<%@ page import="java.net.URLConnection" %>
<%@ page import="java.util.Iterator" %>
<%@ page import="java.util.Map" %>
<%@ page import="java.util.Stack" %>
<%@ page import="java.util.jar.JarEntry" %>
<%@ page import="java.util.jar.JarFile" %>
<%@ page import="org.apache.lucene.search.ScoreDoc" %>
<%@ page import="org.apache.lucene.search.TopScoreDocCollector" %>
<%@ page import="org.apache.lucene.search.IndexSearcher" %>
<%@ page import="org.apache.lucene.index.IndexReader" %>
<%@ page import="org.apache.lucene.util.Version" %>
<%@ page import="org.apache.lucene.analysis.standard.StandardAnalyzer" %>
<%@ page import="org.apache.lucene.queryParser.QueryParser" %>
<%@ page import="org.apache.lucene.search.Query" %>
<%@ page import="org.apache.lucene.store.Directory" %>
<%@ page import="java.net.URLEncoder" %>
<%@ page import="org.tobarsegais.webapp.data.Index" %>
<%@ page import="org.tobarsegais.webapp.data.IndexEntry" %>
<%@ page import="java.util.List" %>
<%@ page import="java.util.ArrayList" %>
<%@ page import="org.tobarsegais.webapp.data.IndexTopic" %>
<%@ page import="org.tobarsegais.webapp.data.Topic" %>
<%@ page import="java.util.AbstractMap" %>
<%@ page import="java.util.Collections" %>
<%@ page import="java.util.Comparator" %>
<%@ page import="org.tobarsegais.webapp.data.IndexSee" %>
<%@ page contentType="text/html;charset=UTF-8" language="java" %>
<html>
<head>
    <meta charset="utf-8"/>
    <title>Help - Tobar Segais platform</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="">
    <meta name="author" content="">

    <!-- The styles -->
    <link href="/css/bootstrap.css" rel="stylesheet">
    <style type="text/css">
        @media print {
            .no-print {
                display: none !important;
            }
        }

        body {
            padding-top: 60px;
            padding-bottom: 40px;
        }

    </style>
    <link href="/css/bootstrap-responsive.css" rel="stylesheet">

    <!-- The HTML5 shim, for IE6-8 support of HTML5 elements -->
    <!--[if lt IE 9]>
    <script src="http://html5shim.googlecode.com/svn/trunk/html5.js"></script>
    <![endif]-->

    <!-- The fav and touch icons -->
    <link rel="shortcut icon" href="/favicon.ico">
    <link rel="apple-touch-icon-precomposed" sizes="144x144" href="/img/apple-touch-icon-144x144.png">
    <link rel="apple-touch-icon-precomposed" sizes="114x114" href="/img/apple-touch-icon-114x114.png">
    <link rel="apple-touch-icon-precomposed" sizes="72x72" href="/img/apple-touch-icon-72x72.png">
    <link rel="apple-touch-icon-precomposed" href="/img/apple-touch-icon.png">
</head>
<body>
<div class="navbar navbar-fixed-top no-print">
    <div class="navbar-inner">
        <div class="container-fluid">
            <a class="btn btn-navbar" data-toggle="collapse" data-target=".nav-collapse">
              <span class="icon-bar"></span>
              <span class="icon-bar"></span>
              <span class="icon-bar"></span>
            </a>
            <a class="brand" href="http://www.tobarsegais.org/">
                <img src="/img/header-logo.png" border="0"/>
                Tobar Segais
            </a>
        </div>
    </div>
</div>
<div class="container-fluid">
    <div class="row-fluid">
        <div class="span4 no-print">
            <div class="well sidebar-nav no-print">
                <%
                    String query = request.getParameter("query");
                    String keywordParam = request.getParameter("keywords");
                    String contentsActive = "";
                    String indexActive = "";
                    String searchActive = "";
                    if (query != null && query.length() > 0) {
                        searchActive = "active";
                    } else if (keywordParam != null) {
                        indexActive = "active";
                    } else {
                        contentsActive = "active";
                    }
                %>
                <ul class="nav nav-tabs">
                    <li class="<%=contentsActive%>"><a href="#contents-nav" data-toggle="tab"><i class="icon-book"></i>Contents</a>
                    </li>
                    <li class="<%=indexActive%>"><a href="#index-nav" data-toggle="tab"><i class="icon-list"></i>Index</a></li>
                    <li class="<%=searchActive%>"><a href="#search-nav" data-toggle="tab"><i class="icon-search"></i>Search</a></li>
                </ul>
                <div class="tab-content" id="sidebar-content">
                    <div class="tab-pane <%=contentsActive%>" id="contents-nav">
                        <%

                            String path = (String) request.getAttribute("content");
                            Map<String, Toc> contents = (Map<String, Toc>) application.getAttribute("toc");

                        %>
                        <ul id="toc">
                                <%
                              for (Map.Entry<String, Toc> bundleEntry : contents.entrySet()) {
                                  TocEntry entry = bundleEntry.getValue();
                                  String bundle = bundleEntry.getKey();
                          %>
                            <li>
                                    <%if (entry.getHref() != null) {%>
                                <a href="/docs/<%=bundle+"/"+entry.getHref()%>">
                                    <%}%>
                              <span
                                      class="<%=entry.getChildren().isEmpty()?"file":"folder"%>"><%=entry
                                      .getLabel()%></span>
                                    <%if (entry.getHref() != null) {%>
                                </a>
                                    <%}%>
                                    <%
                                  Stack<Iterator<? extends TocEntry>> stack = new Stack<Iterator<? extends TocEntry>>();
                                  if (!entry.getChildren().isEmpty()) {
                                      %>
                                <ul><%
                                    stack.push(entry.getChildren().iterator());
                                    while (!stack.empty()) {
                                        Iterator<? extends TocEntry> cur = stack.pop();
                                        if (cur.hasNext()) {
                                            entry = cur.next();
                                            stack.push(cur);
                                %>
                                    <li><%
                                    %>
                                        <%if (entry.getHref() != null) {%>
                                        <a href="/docs/<%=bundle+"/"+entry.getHref()%>">
                                            <%}%>
                              <span
                                      class="<%=entry.getChildren().isEmpty()?"file":"folder"%>"><%=entry
                                      .getLabel()%></span>
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
                    </div>
                    <div class="tab-pane <%=indexActive%>" id="index-nav">
                        <ul>
                            <%
                                Index keywords = (Index) application.getAttribute("keywords");

                                Stack<Iterator<IndexEntry>> stack = new Stack<Iterator<IndexEntry>>();

                                if (keywords != null) {
                                    stack.push(keywords.getEntries().values().iterator());
                                    while (!stack.isEmpty()) {
                                        final Iterator<IndexEntry> iterator = stack.pop();
                                        if (iterator.hasNext()) {
                                            IndexEntry entry = iterator.next();
                                            stack.push(iterator);
                                            out.print("<li id=\"kwdidx-");
                                            out.print(keywords.getId(entry));
                                            out.print("\">");
                            %><%=entry.getKeyword()%><%
                            if (entry.hasChildren()) {
                                out.print("<ul>");
                                List<Map.Entry<String, String>> topics = new ArrayList<Map.Entry<String, String>>(entry.getTopics().size());
                                for (IndexTopic topic : entry.getTopics()) {
                                    String title = topic.getTitle();
                                    if (title == null) {
                                        String href = topic.getHref().replaceAll("#.*$", "");
                                        Toc toc = contents.get(topic.getBundle());
                                        Topic tocTopic = toc.lookupTopic(href);
                                        if (tocTopic != null) {
                                            title = tocTopic.getLabel();
                                        }
                                    }
                                    if (title != null) {
                                        String href = topic.getHref();
                                        if (href.indexOf('#') == -1) {
                                            href = href + "?keywords";
                                        } else {
                                            href = href.replaceFirst("#", "?keywords#");
                                        }
                                        topics.add(new AbstractMap.SimpleImmutableEntry<String, String>(topic.getBundle()+"/"+href, title));
                                    }
                                }
                                Collections.sort(topics, new Comparator<Map.Entry<String, String>>() {
                                    public int compare(Map.Entry<String, String> o1, Map.Entry<String, String> o2) {
                                        return o1.getValue().compareToIgnoreCase(o2.getValue());
                                    }
                                });
                                for (Map.Entry<String, String> topic : topics) {
                        %>
                            <li><a href="/docs/<%=topic.getKey()%>"><%=topic.getValue()%>
                            </a></li>
                            <%
                                }
                                for (IndexSee see : entry.getSees()) {
                                    final IndexEntry indexEntry = keywords.findEntry(see.getKeywordPath());
                                    if (indexEntry != null) {
                                        final String id = keywords.getId(indexEntry);
                                        if (id != null) {
                                            StringBuilder buf = new StringBuilder();
                                            boolean first = true;
                                            for (String keyword : see.getKeywordPath()) {
                                                if (first) {
                                                    first = false;
                                                } else {
                                                    buf.append(", ");
                                                }
                                                buf.append(keyword);
                                            }
                            %>
                            <li>See <a href="#kwdidx-<%=id%>"><%=buf.toString()%>
                            </a></li>
                            <%
                                                        }
                                                    }
                                                }
                                                if (entry.getSubEntries().isEmpty()) {
                                                    out.print("</ul></li>");
                                                } else {
                                                    stack.push(entry.getSubEntries().values().iterator());
                                                }
                                            } else {
                                                out.print("</li>");
                                            }
                                        } else if (!stack.isEmpty()) {
                                            out.print("</ul></li>");
                                        }
                                    }
                                }
                            %>
                        </ul>                    </div>
                    <div class="tab-pane <%=searchActive%>" id="search-nav">
                        <form class="form-search" method="get" action=".">
                            <input name="query" type="query" class="input-medium search-query" value="<%=query==null?"":query%>">
                            <button type="submit" class="btn">Search</button>
                        </form>
                        <%
                            if (query != null && !query.isEmpty()) {
                                Directory index = (Directory) application.getAttribute("index");
                                Query q = new QueryParser(Version.LUCENE_34, "contents", new StandardAnalyzer(Version.LUCENE_34)).parse(query);
                                int hitsPerPage = 200;
                                    IndexReader reader = IndexReader.open(index);
                                    IndexSearcher searcher = new IndexSearcher(reader);
                                    TopScoreDocCollector collector = TopScoreDocCollector.create(hitsPerPage, true);
                                    searcher.search(q, collector);
                                    ScoreDoc[] hits = collector.topDocs().scoreDocs;

                                %>
                        Found <%=hits.length%> hits.
                        <ul>

                        <%
                                    // 4. display results
                                    for(int i=0;i<hits.length;++i) {
                                      int docId = hits[i].doc;
                                      org.apache.lucene.document.Document d = searcher.doc(docId);
                                        String href = d.get("href");
                                        int hashIndex = href.indexOf('#');
                                        String hash = hashIndex == -1 ? "" : href.substring(hashIndex);
                                        href = hashIndex == -1 ? href : href.substring(0, hashIndex);
                                        String url = href + "?query=" + URLEncoder.encode(query, "UTF-8") + hash;
                                        %>
                            <li><a href="/docs/<%=url%>"><%=d.get("title")%></a></li>
                        <%
                                    }

                                    // searcher can only be closed when there
                                    // is no need to access the documents any more.
                                    searcher.close();
                            }
                        %>
                        </ul>
                    </div>
                </div>
            </div>
            <!--/.well -->
        </div>
        <!--/span-->
        <div class="span8 scrollable">
            <div id="content">
            <%
                Map<String, String> bundles = (Map<String, String>) application.getAttribute("bundles");
                for (int index = path.indexOf('/'); index != -1; index = path.indexOf('/', index + 1)) {
                    String key = path.substring(0, index);
                    if (key.startsWith("/")) {
                        key = key.substring(1);
                    }
                    if (bundles.containsKey(key)) {
                        key = bundles.get(key);
                    }
                    URL resource = application.getResource(ServletContextListenerImpl.BUNDLE_PATH + "/" + key + ".jar");
                    if (resource == null) {
                        continue;
                    }
                    URL jarResource = new URL("jar:" + resource + "!/");
                    URLConnection connection = jarResource.openConnection();
                    if (!(connection instanceof JarURLConnection)) {
                        continue;
                    }
                    JarURLConnection jarConnection = (JarURLConnection) connection;
                    JarFile jarFile = jarConnection.getJarFile();

                    int endOfFileName = path.indexOf('#', index);
                    endOfFileName = endOfFileName == -1 ? path.length() : endOfFileName;
                    String fileName = path.substring(index + 1, endOfFileName);
                    JarEntry jarEntry = jarFile.getJarEntry(fileName);
                    if (jarEntry == null) {
                        continue;
                    }
                    InputStream in = null;
                    try {
                        in = jarFile.getInputStream(jarEntry);
                        Document document = Jsoup.parse(in, "UTF-8", request.getRequestURI());
                        out.print(document.body());
                    } finally {
                        IOUtils.closeQuietly(in);
                    }
                }

            %>
        </div>
            </div>
    </div>
</div>
<script src="/js/jquery-latest.js"></script>
<script src="/js/bootstrap.min.js"></script>
<script>
    var TobairSegais = {
        toRawUri: function (uri) {
                var i = uri.indexOf('#');
                if (i != -1) {
                    uri = uri.substring(0, i);
                }
                i = uri.indexOf('?');
                if (i != -1) {
                    uri = uri.substring(0, i);
                }
                return uri + "?raw";
            },
        clickSupport: function() {
            return TobairSegais.loadContent($(this).attr('href'));
        },
        addClickSupport: function(id) {
            $(id + ' a').each(function () {
                $(this).click(TobairSegais.clickSupport);
            });
        },
        loadContent: function(url) {
            if (/^https?:\/\//.test(url)) {
                return true;
            } else {
                $('#content').load(TobairSegais.toRawUri(url), function () {
                    TobairSegais.addClickSupport("#content");
                    history.pushState({url:url}, "", url);
                    var i = url.indexOf('#');
                    if (i != -1) {
                        window.location.href = url.substring(i);
                    } else {
                        window.location.href = "#";
                    }
                });
                return false;
            }
        }
    }
    window.onpopstate = function (event) {
        if (event != null && event.state != null) {
            $('#content').load(TobairSegais.toRawUri(event.state.url), function () {
                TobairSegais.addClickSupport("#content");
            });
        }
    }
    $(function() {
        TobairSegais.addClickSupport("#content");
        TobairSegais.addClickSupport("#sidebar-content");
    });
</script>

</body>
</html>