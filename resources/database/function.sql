--  Copyright (c) 2023. Tus1688
--
--  Permission is hereby granted, free of charge, to any person obtaining a copy
--  of this software and associated documentation files (the "Software"), to deal
--  in the Software without restriction, including without limitation the rights
--  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
--  copies of the Software, and to permit persons to whom the Software is
--  furnished to do so, subject to the following conditions:
--
--  The above copyright notice and this permission notice shall be included in all
--  copies or substantial portions of the Software.
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
--  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
--  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
--  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
--  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
--  SOFTWARE.

CREATE FUNCTION CAP_FIRST(input VARCHAR(304))
    RETURNS VARCHAR(304)
    DETERMINISTIC
BEGIN
    DECLARE len INT;
    DECLARE i INT;
    SET len = CHAR_LENGTH(input);
    SET input = LOWER(input);
    SET i = 0;
    WHILE (i < len)
        DO
            IF (MID(input, i, 1) = ' ' OR i = 0) THEN
                IF (i < len) THEN
                    SET input = CONCAT(
                            LEFT(input, i),
                            UPPER(MID(input, i + 1, 1)),
                            RIGHT(input, len - i - 1)
                        );
                END IF;
            END IF;
            SET i = i + 1;
        END WHILE;
    RETURN input;
END;
