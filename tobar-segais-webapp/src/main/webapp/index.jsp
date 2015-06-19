<%@ page import="org.apache.commons.lang3.StringUtils" %>
<%@ page import="org.tobarsegais.webapp.ServletContextListenerImpl" %>
<%
    String defaultPath = ServletContextListenerImpl.getInitParameter(application, "default.page.path");
    if (StringUtils.isNotBlank(defaultPath)) {
        response.sendRedirect(defaultPath.startsWith("/") ? "docs" + defaultPath : "docs/" + defaultPath);
    } else {
        response.sendRedirect("docs/");
    }
%>
