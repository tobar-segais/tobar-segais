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

<%@ tag trimDirectiveWhitespaces="true" %>
<%@ tag import="org.tobarsegais.webapp.data.Index" %>
<%@ tag import="java.util.Stack" %>
<%@ tag import="java.util.Iterator" %>
<%@ tag import="org.tobarsegais.webapp.data.IndexEntry" %>
<%@ tag import="org.apache.commons.lang3.StringEscapeUtils" %>
<%@ tag import="java.util.List" %>
<%@ tag import="java.util.Map" %>
<%@ tag import="java.util.ArrayList" %>
<%@ tag import="org.tobarsegais.webapp.data.IndexTopic" %>
<%@ tag import="org.tobarsegais.webapp.data.Toc" %>
<%@ tag import="org.tobarsegais.webapp.data.TocEntry" %>
<%@ tag import="java.util.AbstractMap" %>
<%@ tag import="java.util.Collections" %>
<%@ tag import="java.util.Comparator" %>
<%@ tag import="org.tobarsegais.webapp.data.IndexSee" %>
<%@ tag import="java.text.MessageFormat" %>
<%@ tag import="org.tobarsegais.webapp.ServletContextListenerImpl" %>
<ul>
        <%
            Map<String, Toc> contents = ServletContextListenerImpl.getTablesOfContents(application);
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
                        out.print("\"><span>");
                        out.print(StringEscapeUtils.escapeHtml4(entry.getKeyword()));
                        out.print("</span>");
                        if (entry.hasChildren()) {
                            out.print("<ul>");
                            List<Map.Entry<String, String>> topics =
                                    new ArrayList<Map.Entry<String, String>>(entry.getTopics().size());
                            for (IndexTopic topic : entry.getTopics()) {
                                String title = topic.getTitle();
                                if (title == null) {
                                    String href = topic.getHref().replaceAll("#.*$", "");
                                    Toc toc = contents.get(topic.getBundle());
                                    TocEntry tocTopic = toc.lookupTopic(href);
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
                                    topics.add(new AbstractMap.SimpleImmutableEntry<String, String>(
                                            topic.getBundle() + "/" + href,
                                            title));
                                }
                            }
                            Collections.sort(topics, new Comparator<Map.Entry<String, String>>() {
                                public int compare(Map.Entry<String, String> o1, Map.Entry<String, String> o2) {
                                    return o1.getValue().compareToIgnoreCase(o2.getValue());
                                }
                            });
                            for (Map.Entry<String, String> topic : topics) {
                                out.print("<li><a href=\"");
                                out.print(request.getContextPath());
                                out.print("/docs/");
                                out.print(topic.getKey());
                                out.print("\">");
                                out.print(StringEscapeUtils.escapeHtml4(topic.getValue()));
                                out.print("</a></li>");
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
                                        out.print("<li>");
                                        final String linkHtml =
                                                MessageFormat.format("<a href=\"{0}\" ts-immediate=\"true\">{1}</a>",
                                                        "#kwdidx-" + id, StringEscapeUtils.escapeHtml4(buf.toString()));
                                        out.print(MessageFormat.format("See {0}", linkHtml));
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
    </ul>
