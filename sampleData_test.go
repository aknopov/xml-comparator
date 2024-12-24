package xmlcomparator

var xmlString1 = `
<note color="red">
    <to>Tove</to>
    <from>Jani</from>
    <date>2023-08-27T16:27:55+00:00</date>
    <heading>Reminder</heading>
    <body>Don't forget me this weekend!</body>
</note>`

var xmlString2 = `
<root type="vet_hospital">
	<animal>
		<p>This is dog</p>
		<dog>
			<p>tommy</p>
		</dog>
	</animal>
	<birds>
		<p class="bar">this is birds</p>
		<p>this is birds</p>
	</birds>
	<animal>
		<p>this is animals</p>
	</animal>
</root>`

var xmlMixed = `
<note color="red">
    Some text ...
    <to>Tove</to>
	mixed with elements
    <from>Jani</from>
    <date>2023-08-27T16:27:55+00:00</date>
    <heading>Reminder</heading>
    <body>Don't forget me this weekend!</body>
</note>`

var soapString = `
<?xml version = "1.0"?>
<SOAP-ENV:Envelope
	xmlns:SOAP-ENV = "http://www.w3.org/2001/12/soap-envelope"
	SOAP-ENV:encodingStyle = "http://www.w3.org/2001/12/soap-encoding">
	<SOAP-ENV:Body xmlns:m = "http://www.xyz.org/quotations">
		<m:GetQuotation>
			<m:QuotationsName>MiscroSoft</m:QuotationsName>
		</m:GetQuotation>
	</SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
