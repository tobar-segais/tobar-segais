package org.tobarsegais.webapp;

import org.apache.commons.io.IOUtils;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.URL;

/**
 * Created by IntelliJ IDEA.
 * User: stephenc
 * Date: 18/10/2011
 * Time: 09:35
 * To change this template use File | Settings | File Templates.
 */
public class ContentServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws ServletException, IOException {
        String path = req.getPathInfo();
        int index = path.indexOf('/');
        while (index != -1) {
            String realPath = getServletContext().getRealPath("/WEB-INF/bundles" + path.substring(0, index) + ".jar");
            File file = new File(realPath);
            if (file.isFile()) {
                int eofn = path.indexOf('#', index);
                eofn = eofn == -1 ? path.length() : eofn;
                String fileName = path.substring(index, eofn);
                URL url = new URL("jar:file://" + file + "!" + fileName);
                InputStream in = null;
                OutputStream out = resp.getOutputStream();
                try {
                    in = url.openStream();
                    IOUtils.copy(in, out);
                } finally {
                    IOUtils.closeQuietly(in);
                    out.close();
                }
                return;
            }
            index = path.indexOf('/', index + 1);
        }
        resp.sendError(404);
    }
}
