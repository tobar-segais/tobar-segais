<%@ page import="org.apache.commons.lang3.StringUtils" %>
<%
    String defaultPath = application.getInitParameter("default.page.path");
    if (StringUtils.isNotBlank(defaultPath)) {
        response.sendRedirect(defaultPath.startsWith("/") ? "docs" + defaultPath : "docs/" + defaultPath);
    } else {
        response.sendRedirect("docs");
    }
%>
